package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	// Common settings
	Mode     string `json:"mode"`     // "uploader" or "watcher"
	LogLevel string `json:"log_level"` // "debug", "info", "warn", "error"
	
	// Uploader settings
	Uploader UploaderConfig `json:"uploader"`
	
	// Watcher settings
	Watcher WatcherConfig `json:"watcher"`
}

type UploaderConfig struct {
	DockerBuildPath    string `json:"docker_build_path"`    // Path to Dockerfile
	ImageName          string `json:"image_name"`           // Docker image name
	ImageTag           string `json:"image_tag"`            // Docker image tag
	TarballPath        string `json:"tarball_path"`         // Local path to save tarball
	RemoteHost         string `json:"remote_host"`          // SSH host
	RemotePort         int    `json:"remote_port"`          // SSH port
	RemoteUser         string `json:"remote_user"`          // SSH username
	RemoteKeyPath      string `json:"remote_key_path"`      // SSH private key path
	RemoteUploadPath   string `json:"remote_upload_path"`   // Remote upload directory
	BuildCommand       string `json:"build_command"`        // Custom build command (optional)
	PreBuildCommands   []string `json:"pre_build_commands"` // Commands before build
	PostBuildCommands  []string `json:"post_build_commands"`// Commands after build
}

type WatcherConfig struct {
	WatchDirectory     string `json:"watch_directory"`      // Directory to watch for tarballs
	ContainerName      string `json:"container_name"`       // Container name to manage
	ContainerPort      []string `json:"container_ports"`    // Port mappings
	ContainerEnv       []string `json:"container_env"`      // Environment variables
	ContainerVolumes   []string `json:"container_volumes"`  // Volume mappings
	PreLoadCommands    []string `json:"pre_load_commands"`  // Commands before loading image
	PostLoadCommands   []string `json:"post_load_commands"` // Commands after loading image
	RestartPolicy      string   `json:"restart_policy"`     // Docker restart policy
}

func LoadConfig(configPath string) (*Config, error) {
	config := &Config{
		Mode:     "watcher",
		LogLevel: "info",
		Uploader: UploaderConfig{
			RemotePort: 22,
			ImageTag:   "latest",
		},
		Watcher: WatcherConfig{
			RestartPolicy: "unless-stopped",
		},
	}

	if configPath == "" {
		return config, nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return config, nil
}

func (c *Config) SaveConfig(configPath string) error {
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config file: %w", err)
	}

	return nil
}

func (c *Config) Validate() error {
	if c.Mode != "uploader" && c.Mode != "watcher" {
		return fmt.Errorf("invalid mode: %s (must be 'uploader' or 'watcher')", c.Mode)
	}

	if c.Mode == "uploader" {
		if c.Uploader.DockerBuildPath == "" {
			return fmt.Errorf("docker_build_path is required for uploader mode")
		}
		if c.Uploader.ImageName == "" {
			return fmt.Errorf("image_name is required for uploader mode")
		}
		if c.Uploader.RemoteHost == "" {
			return fmt.Errorf("remote_host is required for uploader mode")
		}
		if c.Uploader.RemoteUser == "" {
			return fmt.Errorf("remote_user is required for uploader mode")
		}
		if c.Uploader.RemoteUploadPath == "" {
			return fmt.Errorf("remote_upload_path is required for uploader mode")
		}
	}

	if c.Mode == "watcher" {
		if c.Watcher.WatchDirectory == "" {
			return fmt.Errorf("watch_directory is required for watcher mode")
		}
		if c.Watcher.ContainerName == "" {
			return fmt.Errorf("container_name is required for watcher mode")
		}
	}

	return nil
} 