package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/umarahsan/fws/internal/config"
	"github.com/umarahsan/fws/internal/uploader"
	"github.com/umarahsan/fws/internal/utils"
	"github.com/umarahsan/fws/internal/watcher"
)

var (
	configFile string
	mode       string
	daemon     bool
	verbose    bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "file-watch-server",
	Short: "A file watcher and uploader for Docker containers",
	Long: `File Watch Server is a tool that operates in two modes:

1. Uploader Mode: Builds Docker images, creates tarballs, and uploads them via SSH/SCP
2. Watcher Mode: Watches for uploaded tarballs and automatically deploys them as containers

The application can run as a daemon in the background on any Linux platform.`,
	Run: func(cmd *cobra.Command, args []string) {
		runApplication()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path (default: ./config.json)")
	rootCmd.PersistentFlags().StringVarP(&mode, "mode", "m", "", "operation mode: uploader or watcher")
	rootCmd.PersistentFlags().BoolVarP(&daemon, "daemon", "d", false, "run as daemon in background")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output (debug level)")
}

func runApplication() {
	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Override config with command line flags
	if mode != "" {
		cfg.Mode = mode
	}
	if verbose {
		cfg.LogLevel = "debug"
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// Create logger
	logger := utils.NewLogger(cfg.LogLevel)

	// Run based on mode
	switch cfg.Mode {
	case "uploader":
		runUploader(cfg, logger)
	case "watcher":
		runWatcher(cfg, logger, daemon)
	default:
		fmt.Printf("Invalid mode: %s. Must be 'uploader' or 'watcher'\n", cfg.Mode)
		os.Exit(1)
	}
}

func runUploader(cfg *config.Config, logger *utils.Logger) {
	logger.Info("Starting in uploader mode...")
	
	up := uploader.NewUploader(&cfg.Uploader, logger)
	if err := up.Run(); err != nil {
		logger.Fatal("Uploader failed: %v", err)
	}
}

func runWatcher(cfg *config.Config, logger *utils.Logger, isDaemon bool) {
	logger.Info("Starting in watcher mode...")
	
	w := watcher.NewWatcher(&cfg.Watcher, logger)
	
	if isDaemon {
		// Run as daemon
		err := utils.Daemonize(func() error {
			return runWatcherWithSignalHandling(w, logger)
		}, logger)
		if err != nil {
			logger.Fatal("Failed to start daemon: %v", err)
		}
	} else {
		// Run in foreground
		if err := runWatcherWithSignalHandling(w, logger); err != nil {
			logger.Fatal("Watcher failed: %v", err)
		}
	}
}

func runWatcherWithSignalHandling(w *watcher.Watcher, logger *utils.Logger) error {
	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start watcher in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- w.Run()
	}()

	// Wait for signal or error
	select {
	case sig := <-sigChan:
		logger.Info("Received signal: %v", sig)
		w.Stop()
		return nil
	case err := <-errChan:
		return err
	}
}

// Additional commands for configuration management
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long:  `Create a sample configuration file with default values.`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show container status (watcher mode only)",
	Long:  `Display the status of the managed container.`,
	Run: func(cmd *cobra.Command, args []string) {
		showStatus()
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show container logs (watcher mode only)",
	Long:  `Display the logs of the managed container.`,
	Run: func(cmd *cobra.Command, args []string) {
		showLogs()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
}

func initConfig() {
	configPath := "config.json"
	if configFile != "" {
		configPath = configFile
	}

	// Create default config
	cfg := &config.Config{
		Mode:     "watcher",
		LogLevel: "info",
		Uploader: config.UploaderConfig{
			DockerBuildPath:  "./",
			ImageName:        "myapp",
			ImageTag:         "latest",
			TarballPath:      "./tarballs",
			RemoteHost:       "destination.server.com",
			RemotePort:       22,
			RemoteUser:       "deploy",
			RemoteKeyPath:    "~/.ssh/id_rsa",
			RemoteUploadPath: "/opt/docker-uploads",
			PreBuildCommands: []string{
				"echo 'Starting build process...'",
			},
			PostBuildCommands: []string{
				"echo 'Build process completed.'",
			},
		},
		Watcher: config.WatcherConfig{
			WatchDirectory:   "/opt/docker-uploads",
			ContainerName:    "myapp",
			ContainerPort:    []string{"8080:8080"},
			ContainerEnv:     []string{"NODE_ENV=production"},
			ContainerVolumes: []string{},
			PreLoadCommands: []string{
				"echo 'Preparing to load new image...'",
			},
			PostLoadCommands: []string{
				"echo 'New container deployed successfully.'",
			},
			RestartPolicy: "unless-stopped",
		},
	}

	if err := cfg.SaveConfig(configPath); err != nil {
		fmt.Printf("Failed to save config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Configuration file created: %s\n", configPath)
	fmt.Println("Please edit the configuration file before running the application.")
}

func showStatus() {
	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if cfg.Mode != "watcher" {
		fmt.Println("Status command is only available in watcher mode")
		os.Exit(1)
	}

	logger := utils.NewLogger(cfg.LogLevel)
	w := watcher.NewWatcher(&cfg.Watcher, logger)

	status, err := w.GetContainerStatus()
	if err != nil {
		fmt.Printf("Failed to get container status: %v\n", err)
		os.Exit(1)
	}

	if status == "" {
		fmt.Printf("Container '%s' not found\n", cfg.Watcher.ContainerName)
	} else {
		fmt.Printf("Container '%s' status: %s\n", cfg.Watcher.ContainerName, status)
	}
}

func showLogs() {
	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if cfg.Mode != "watcher" {
		fmt.Println("Logs command is only available in watcher mode")
		os.Exit(1)
	}

	logger := utils.NewLogger(cfg.LogLevel)
	w := watcher.NewWatcher(&cfg.Watcher, logger)

	logs, err := w.GetContainerLogs(50)
	if err != nil {
		fmt.Printf("Failed to get container logs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Container '%s' logs:\n", cfg.Watcher.ContainerName)
	fmt.Println(logs)
} 