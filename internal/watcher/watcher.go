package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/umarahsan/fws/internal/config"
	"github.com/umarahsan/fws/internal/utils"
)

type Watcher struct {
	config  *config.WatcherConfig
	logger  *utils.Logger
	watcher *fsnotify.Watcher
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewWatcher(cfg *config.WatcherConfig, logger *utils.Logger) *Watcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &Watcher{
		config: cfg,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Run starts the file watcher daemon
func (w *Watcher) Run() error {
	w.logger.Info("Starting file watcher daemon...")

	// Ensure watch directory exists
	if err := utils.EnsureDir(w.config.WatchDirectory); err != nil {
		return fmt.Errorf("failed to create watch directory: %w", err)
	}

	// Create file system watcher
	var err error
	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer w.watcher.Close()

	// Add directory to watch
	if err := w.watcher.Add(w.config.WatchDirectory); err != nil {
		return fmt.Errorf("failed to add directory to watch: %w", err)
	}

	w.logger.Info("Watching directory: %s", w.config.WatchDirectory)

	// Start processing events
	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("File watcher stopped")
			return nil
		case event, ok := <-w.watcher.Events:
			if !ok {
				return fmt.Errorf("file watcher events channel closed")
			}
			w.handleFileEvent(event)
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return fmt.Errorf("file watcher errors channel closed")
			}
			w.logger.Error("File watcher error: %v", err)
		}
	}
}

// Stop stops the file watcher daemon
func (w *Watcher) Stop() {
	w.logger.Info("Stopping file watcher daemon...")
	w.cancel()
}

func (w *Watcher) handleFileEvent(event fsnotify.Event) {
	// Only process .tar files
	if !strings.HasSuffix(event.Name, ".tar") {
		return
	}

	w.logger.Debug("File event: %s %s", event.Op, event.Name)

	// Handle file creation and write events
	if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
		w.logger.Info("New tarball detected: %s", event.Name)

		// Wait a bit to ensure file is fully written
		time.Sleep(2 * time.Second)

		// Process the tarball
		if err := w.processTarball(event.Name); err != nil {
			w.logger.Error("Failed to process tarball %s: %v", event.Name, err)
		}
	}
}

func (w *Watcher) processTarball(tarballPath string) error {
	w.logger.Info("Processing tarball: %s", tarballPath)

	// Check if file exists and is readable
	if !utils.FileExists(tarballPath) {
		return fmt.Errorf("tarball does not exist: %s", tarballPath)
	}

	// Get file size for logging
	size, err := utils.GetFileSize(tarballPath)
	if err != nil {
		w.logger.Warn("Failed to get tarball size: %v", err)
	} else {
		w.logger.Info("Processing tarball: %s (%s)", filepath.Base(tarballPath), utils.FormatBytes(size))
	}

	// Execute pre-load commands
	if err := w.executePreLoadCommands(); err != nil {
		return fmt.Errorf("pre-load commands failed: %w", err)
	}

	// Load Docker image from tarball
	if err := w.loadDockerImage(tarballPath); err != nil {
		return fmt.Errorf("failed to load Docker image: %w", err)
	}

	// Stop and remove existing container
	if err := w.stopAndRemoveContainer(); err != nil {
		w.logger.Warn("Failed to stop/remove existing container: %v", err)
	}

	// Start new container
	if err := w.startContainer(); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Execute post-load commands
	if err := w.executePostLoadCommands(); err != nil {
		w.logger.Warn("Post-load commands failed: %v", err)
	}

	// Clean up tarball
	if err := w.cleanupTarball(tarballPath); err != nil {
		w.logger.Warn("Failed to cleanup tarball: %v", err)
	}

	w.logger.Info("Tarball processing completed successfully")
	return nil
}

func (w *Watcher) executePreLoadCommands() error {
	if len(w.config.PreLoadCommands) == 0 {
		return nil
	}

	w.logger.Info("Executing pre-load commands...")
	return utils.ExecuteCommands(w.config.PreLoadCommands, 5*time.Minute, w.logger)
}

func (w *Watcher) loadDockerImage(tarballPath string) error {
	w.logger.Info("Loading Docker image from tarball: %s", tarballPath)

	loadCmd := fmt.Sprintf("docker load -i %s", tarballPath)
	output, err := utils.ExecuteCommand(loadCmd, 10*time.Minute)
	if err != nil {
		return err
	}

	w.logger.Debug("Docker load output: %s", strings.TrimSpace(output))
	return nil
}

func (w *Watcher) stopAndRemoveContainer() error {
	w.logger.Info("Stopping and removing existing container: %s", w.config.ContainerName)

	// Stop container
	stopCmd := fmt.Sprintf("docker stop %s", w.config.ContainerName)
	output, err := utils.ExecuteCommand(stopCmd, 30*time.Second)
	if err != nil {
		w.logger.Debug("Failed to stop container (may not exist): %v", err)
	} else {
		w.logger.Debug("Docker stop output: %s", strings.TrimSpace(output))
	}

	// Remove container
	removeCmd := fmt.Sprintf("docker rm %s", w.config.ContainerName)
	output, err = utils.ExecuteCommand(removeCmd, 30*time.Second)
	if err != nil {
		w.logger.Debug("Failed to remove container (may not exist): %v", err)
	} else {
		w.logger.Debug("Docker remove output: %s", strings.TrimSpace(output))
	}

	return nil
}

func (w *Watcher) startContainer() error {
	w.logger.Info("Starting new container: %s", w.config.ContainerName)

	// Build docker run command
	runCmd := w.buildDockerRunCommand()

	output, err := utils.ExecuteCommand(runCmd, 2*time.Minute)
	if err != nil {
		return err
	}

	w.logger.Debug("Docker run output: %s", strings.TrimSpace(output))
	w.logger.Info("Container started successfully: %s", w.config.ContainerName)
	return nil
}

func (w *Watcher) buildDockerRunCommand() string {
	var cmd strings.Builder
	cmd.WriteString("docker run -d")

	// Add container name
	cmd.WriteString(fmt.Sprintf(" --name %s", w.config.ContainerName))

	// Add restart policy
	if w.config.RestartPolicy != "" {
		cmd.WriteString(fmt.Sprintf(" --restart %s", w.config.RestartPolicy))
	}

	// Add port mappings
	for _, port := range w.config.ContainerPort {
		cmd.WriteString(fmt.Sprintf(" -p %s", port))
	}

	// Add environment variables
	for _, env := range w.config.ContainerEnv {
		cmd.WriteString(fmt.Sprintf(" -e %s", env))
	}

	// Add volume mappings
	for _, volume := range w.config.ContainerVolumes {
		cmd.WriteString(fmt.Sprintf(" -v %s", volume))
	}

	// Extract image name from tarball filename
	imageName := w.extractImageNameFromTarball()
	cmd.WriteString(fmt.Sprintf(" %s", imageName))

	return cmd.String()
}

func (w *Watcher) extractImageNameFromTarball() string {
	// This is a simplified version - in a real implementation, you might want to
	// parse the tarball or maintain a mapping of tarball names to image names
	// For now, we'll assume the image name is derived from the container name
	return w.config.ContainerName
}

func (w *Watcher) executePostLoadCommands() error {
	if len(w.config.PostLoadCommands) == 0 {
		return nil
	}

	w.logger.Info("Executing post-load commands...")
	return utils.ExecuteCommands(w.config.PostLoadCommands, 5*time.Minute, w.logger)
}

func (w *Watcher) cleanupTarball(tarballPath string) error {
	w.logger.Info("Cleaning up tarball: %s", tarballPath)
	return os.Remove(tarballPath)
}

// GetContainerStatus returns the status of the managed container
func (w *Watcher) GetContainerStatus() (string, error) {
	statusCmd := fmt.Sprintf("docker ps -a --filter name=%s --format '{{.Status}}'", w.config.ContainerName)
	output, err := utils.ExecuteCommand(statusCmd, 10*time.Second)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// GetContainerLogs returns the logs of the managed container
func (w *Watcher) GetContainerLogs(lines int) (string, error) {
	logsCmd := fmt.Sprintf("docker logs --tail %d %s", lines, w.config.ContainerName)
	output, err := utils.ExecuteCommand(logsCmd, 30*time.Second)
	if err != nil {
		return "", err
	}

	return output, nil
}
