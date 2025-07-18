package providers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

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

	// Execute claude CLI with --print flag for non-interactive output
	args := p.buildArgs("--print")
	cmd := exec.CommandContext(ctx, p.cliPath, args...)
	cmd.Stdin = bytes.NewReader([]byte(prompt))
	
	// Inherit environment variables including PATH and HOME for Claude auth
	// Add environment variables to prevent TTY issues in Docker
	cmd.Env = append(os.Environ(), 
		"CI=true",                    // Prevent interactive prompts
		"TERM=dumb",                  // Simple terminal
		"NO_COLOR=1",                 // Disable colors
		"CLAUDE_DISABLE_RAW_MODE=1",  // Disable raw mode
	)
	
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

// StreamResponse streams Claude CLI response to the provided writer
func (p *ClaudeProvider) StreamResponse(ctx context.Context, prompt string, chatID int64, writer io.Writer) error {
	// Setup logging
	logFile, err := p.setupLogging(chatID, prompt)
	if err != nil {
		return err
	}
	defer logFile.Close()

	// Prepare temporary file for prompt
	tmpFileName, cleanup, err := p.createTempPromptFile(prompt)
	if err != nil {
		return err
	}
	defer cleanup()

	// Setup and start Claude CLI command
	cmd, stdout, stderr, err := p.setupClaudeCommand(ctx, tmpFileName)
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude CLI: %w", err)
	}

	// Handle command execution and output
	return p.handleCommandExecution(cmd, stdout, stderr, writer, logFile)
}

// setupLogging creates and initializes the log file for the chat
func (p *ClaudeProvider) setupLogging(chatID int64, prompt string) (*os.File, error) {
	logPath := fmt.Sprintf("%s/claude/chat_%d.log", p.logDir, chatID)
	logFile, err := utils.CreateFile(logPath)
	if err != nil {
		return nil, err
	}

	// Log the prompt
	fmt.Fprintf(logFile, "USER: %s\n", prompt)
	fmt.Fprintf(logFile, "ASSISTANT: ")

	return logFile, nil
}

// createTempPromptFile creates a temporary file with the prompt content
func (p *ClaudeProvider) createTempPromptFile(prompt string) (string, func(), error) {
	tmpFile, err := os.CreateTemp("", "claude-prompt-*.txt")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpFileName := tmpFile.Name()

	// Cleanup function
	cleanup := func() {
		tmpFile.Close()
		os.Remove(tmpFileName)
	}

	if _, err := tmpFile.WriteString(prompt); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to write prompt to temp file: %w", err)
	}
	tmpFile.Close()

	return tmpFileName, cleanup, nil
}

// setupClaudeCommand creates and configures the Claude CLI command
func (p *ClaudeProvider) setupClaudeCommand(ctx context.Context, tmpFileName string) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	// Build command arguments
	args := p.buildArgs("--print")
	cmd := exec.CommandContext(ctx, p.cliPath, args...)

	// Set stdin to read from temp file
	tmpFileForRead, err := os.Open(tmpFileName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open temp file for reading: %w", err)
	}
	cmd.Stdin = tmpFileForRead

	// Set environment variables to prevent TTY issues
	cmd.Env = append(os.Environ(),
		"CI=true",
		"TERM=dumb",
		"NO_COLOR=1",
		"FORCE_COLOR=0",
	)

	// Get stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	return cmd, stdout, stderr, nil
}

// handleCommandExecution manages the execution and output handling of the Claude CLI command
func (p *ClaudeProvider) handleCommandExecution(cmd *exec.Cmd, stdout, stderr io.ReadCloser, writer io.Writer, logFile *os.File) error {
	// Ensure stdout and stderr are closed properly
	defer stdout.Close()
	defer stderr.Close()
	
	// Close stdin file if it exists
	if cmd.Stdin != nil {
		if file, ok := cmd.Stdin.(*os.File); ok {
			defer file.Close()
		}
	}
	
	// Handle stderr with proper error handling and synchronization
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		p.handleStderr(stderr, logFile)
	}()

	// Create multi-writer to write to both output and log
	multiWriter := io.MultiWriter(writer, logFile)

	// Copy output
	if _, err := io.Copy(multiWriter, stdout); err != nil {
		return fmt.Errorf("failed to copy output: %w", err)
	}

	// Wait for stderr goroutine to complete
	wg.Wait()

	// Add newline to log
	fmt.Fprintf(logFile, "\n")

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("claude CLI failed: %w", err)
	}

	return nil
}

// handleStderr processes stderr output from the Claude CLI command
func (p *ClaudeProvider) handleStderr(stderr io.ReadCloser, logFile *os.File) {
	stderrBytes, err := io.ReadAll(stderr)
	if err != nil {
		utils.Error("Claude CLI stderr read error: %v", err)
		return
	}
	if len(stderrBytes) > 0 {
		utils.Error("Claude CLI stderr: %s", string(stderrBytes))
		fmt.Fprintf(logFile, "\nERROR: %s\n", string(stderrBytes))
	}
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