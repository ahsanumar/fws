package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Logger struct {
	level string
}

func NewLogger(level string) *Logger {
	return &Logger{level: level}
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.level == "debug" {
		log.Printf("[DEBUG] "+msg, args...)
	}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	if l.level == "debug" || l.level == "info" {
		log.Printf("[INFO] "+msg, args...)
	}
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	if l.level == "debug" || l.level == "info" || l.level == "warn" {
		log.Printf("[WARN] "+msg, args...)
	}
}

func (l *Logger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}

func (l *Logger) Fatal(msg string, args ...interface{}) {
	log.Fatalf("[FATAL] "+msg, args...)
}

// ExecuteCommand executes a shell command with timeout
func ExecuteCommand(command string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, err := cmd.CombinedOutput()
	
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("command timed out after %v: %s", timeout, command)
	}
	
	if err != nil {
		return string(output), fmt.Errorf("command failed: %s, output: %s", err.Error(), string(output))
	}
	
	return string(output), nil
}

// ExecuteCommands executes multiple shell commands sequentially
func ExecuteCommands(commands []string, timeout time.Duration, logger *Logger) error {
	for _, cmd := range commands {
		if strings.TrimSpace(cmd) == "" {
			continue
		}
		
		logger.Info("Executing command: %s", cmd)
		output, err := ExecuteCommand(cmd, timeout)
		
		if err != nil {
			logger.Error("Command failed: %s", err.Error())
			return err
		}
		
		if strings.TrimSpace(output) != "" {
			logger.Debug("Command output: %s", strings.TrimSpace(output))
		}
	}
	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// EnsureDir ensures a directory exists
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// FormatBytes formats bytes to human readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetTimestamp returns current timestamp in RFC3339 format
func GetTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// Daemonize runs the function in background as a daemon
func Daemonize(fn func() error, logger *Logger) error {
	logger.Info("Starting daemon process...")
	
	// Setup signal handling for graceful shutdown
	// This is a simplified version - you might want to use a proper signal handling library
	
	return fn()
} 