package providers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"ai-gateway-hub/internal/utils"
)

// ClaudeProvider implements the AIProvider interface for Claude CLI
type ClaudeProvider struct {
	cliPath         string
	logDir          string
	skipPermissions bool
	extraArgs       string
}

// NewClaudeProvider creates a new Claude provider instance
func NewClaudeProvider(cliPath, logDir string, skipPermissions bool, extraArgs string) *ClaudeProvider {
	return &ClaudeProvider{
		cliPath:         cliPath,
		logDir:          logDir,
		skipPermissions: skipPermissions,
		extraArgs:       extraArgs,
	}
}

func (p *ClaudeProvider) GetID() string {
	return "claude"
}

func (p *ClaudeProvider) GetName() string {
	return "Claude Code"
}

func (p *ClaudeProvider) GetDescription() string {
	return "Anthropic's Claude AI assistant via CLI"
}

func (p *ClaudeProvider) IsAvailable() bool {
	// Check if claude CLI is available
	cmd := exec.Command(p.cliPath, "--version")
	cmd.Env = os.Environ()
	err := cmd.Run()
	return err == nil
}

func (p *ClaudeProvider) GetStatus() ProviderStatus {
	status := ProviderStatus{
		Available: false,
		Status:    "not_installed",
		Details:   "Claude CLI not found",
	}

	// Check if claude CLI exists with a quick version check only
	cmd := exec.Command(p.cliPath, "--version")
	cmd.Env = os.Environ()
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if this is a "command not found" error
		if execErr, ok := err.(*exec.Error); ok && execErr.Err == exec.ErrNotFound {
			// Command not found
			status.Status = "not_installed"
			status.Details = fmt.Sprintf("Claude CLI not found at '%s'", p.cliPath)
		} else if strings.Contains(err.Error(), "no such file or directory") || 
		          strings.Contains(err.Error(), "command not found") {
			// Alternative check for command not found
			status.Status = "not_installed"
			status.Details = fmt.Sprintf("Claude CLI not found at '%s'", p.cliPath)
		} else {
			// Command failed for other reasons
			status.Status = "error"
			status.Details = fmt.Sprintf("Claude CLI error: %v", err)
		}
		return status
	}

	// Parse version from output
	version := strings.TrimSpace(string(output))
	status.Version = version

	// If version check succeeded, assume it's ready (skip the help command for performance)
	status.Available = true
	status.Status = "ready"
	status.Details = "Claude CLI is available"
	
	return status
}

// buildArgs constructs the command arguments based on provider configuration
func (p *ClaudeProvider) buildArgs(baseArgs ...string) []string {
	args := make([]string, 0)
	
	// Add base arguments
	args = append(args, baseArgs...)
	
	// Add skip permissions flag if enabled
	if p.skipPermissions {
		args = append(args, "--dangerously-skip-permissions")
	}
	
	// Add extra arguments if provided
	if p.extraArgs != "" {
		// Split extra args by space, respecting quoted strings
		extraArgsList := strings.Fields(p.extraArgs)
		args = append(args, extraArgsList...)
	}
	
	return args
}

func (p *ClaudeProvider) SendPrompt(ctx context.Context, prompt string, chatID int64) (io.ReadCloser, error) {
	// Create log file for this chat
	logPath := fmt.Sprintf("%s/claude/chat_%d.log", p.logDir, chatID)
	logFile, err := utils.CreateFile(logPath)
	if err != nil {
		return nil, err
	}
	defer logFile.Close()

	// Execute claude CLI
	args := p.buildArgs("chat", "--no-stream")
	cmd := exec.CommandContext(ctx, p.cliPath, args...)
	cmd.Stdin = bytes.NewReader([]byte(prompt))
	
	// Inherit environment variables including PATH and HOME for Claude auth
	cmd.Env = os.Environ()
	
	// Log the prompt
	fmt.Fprintf(logFile, "USER: %s\n", prompt)

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start claude CLI: %w", err)
	}

	// Return a reader that logs the response
	return &loggingReader{
		reader:  stdout,
		logFile: logFile,
		cmd:     cmd,
	}, nil
}

func (p *ClaudeProvider) StreamResponse(ctx context.Context, prompt string, chatID int64, writer io.Writer) error {
	// Create log file for this chat
	logPath := fmt.Sprintf("%s/claude/chat_%d.log", p.logDir, chatID)
	logFile, err := utils.CreateFile(logPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	// Execute claude CLI with streaming
	args := p.buildArgs("chat")
	cmd := exec.CommandContext(ctx, p.cliPath, args...)
	cmd.Stdin = bytes.NewReader([]byte(prompt))
	
	// Inherit environment variables including PATH and HOME for Claude auth
	cmd.Env = os.Environ()
	
	// Log the prompt
	fmt.Fprintf(logFile, "USER: %s\n", prompt)
	fmt.Fprintf(logFile, "ASSISTANT: ")

	// Get stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude CLI: %w", err)
	}
	
	// Log any errors
	go func() {
		stderrBytes, _ := io.ReadAll(stderr)
		if len(stderrBytes) > 0 {
			utils.Error("Claude CLI stderr: %s", string(stderrBytes))
			fmt.Fprintf(logFile, "\nERROR: %s\n", string(stderrBytes))
		}
	}()

	// Create multi-writer to write to both output and log
	multiWriter := io.MultiWriter(writer, logFile)

	// Copy output
	if _, err := io.Copy(multiWriter, stdout); err != nil {
		return fmt.Errorf("failed to copy output: %w", err)
	}

	// Add newline to log
	fmt.Fprintf(logFile, "\n")

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("claude CLI failed: %w", err)
	}

	return nil
}

// loggingReader wraps a reader and logs its output
type loggingReader struct {
	reader  io.Reader
	logFile *os.File
	cmd     *exec.Cmd
	buffer  []byte
}

func (lr *loggingReader) Read(p []byte) (n int, err error) {
	n, err = lr.reader.Read(p)
	if n > 0 {
		// Log the response
		lr.buffer = append(lr.buffer, p[:n]...)
	}
	return n, err
}

func (lr *loggingReader) Close() error {
	// Write the complete response to log
	if len(lr.buffer) > 0 {
		fmt.Fprintf(lr.logFile, "ASSISTANT: %s\n", string(lr.buffer))
	}
	
	// Wait for command to finish
	if lr.cmd != nil {
		if err := lr.cmd.Wait(); err != nil {
			utils.Error("Claude CLI wait error: %v", err)
		}
	}
	
	return nil
}