package providers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// ClaudeProvider implements the AIProvider interface for Claude CLI
type ClaudeProvider struct {
	cliPath string
	logDir  string
}

// NewClaudeProvider creates a new Claude provider instance
func NewClaudeProvider(cliPath, logDir string) *ClaudeProvider {
	return &ClaudeProvider{
		cliPath: cliPath,
		logDir:  logDir,
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
	err := cmd.Run()
	return err == nil
}

func (p *ClaudeProvider) SendPrompt(ctx context.Context, prompt string, chatID int64) (io.ReadCloser, error) {
	// Create log file for this chat
	logPath := filepath.Join(p.logDir, "claude", fmt.Sprintf("chat_%d.log", chatID))
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()

	// Execute claude CLI
	cmd := exec.CommandContext(ctx, p.cliPath, "chat", "--no-stream")
	cmd.Stdin = bytes.NewReader([]byte(prompt))
	
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
	logPath := filepath.Join(p.logDir, "claude", fmt.Sprintf("chat_%d.log", chatID))
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()

	// Execute claude CLI with streaming
	cmd := exec.CommandContext(ctx, p.cliPath, "chat")
	cmd.Stdin = bytes.NewReader([]byte(prompt))
	
	// Log the prompt
	fmt.Fprintf(logFile, "USER: %s\n", prompt)
	fmt.Fprintf(logFile, "ASSISTANT: ")

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude CLI: %w", err)
	}

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
			log.Printf("Claude CLI wait error: %v", err)
		}
	}
	
	return nil
}