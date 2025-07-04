package uploader

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/umarahsan/fws/internal/config"
	"github.com/umarahsan/fws/internal/utils"
)

type Uploader struct {
	config *config.UploaderConfig
	logger *utils.Logger
}

func NewUploader(cfg *config.UploaderConfig, logger *utils.Logger) *Uploader {
	return &Uploader{
		config: cfg,
		logger: logger,
	}
}

// Run executes the full uploader workflow
func (u *Uploader) Run() error {
	u.logger.Info("Starting uploader workflow...")

	// Execute pre-build commands
	if err := u.executePreBuildCommands(); err != nil {
		return fmt.Errorf("pre-build commands failed: %w", err)
	}

	// Build Docker image
	if err := u.buildDockerImage(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	// Create tarball
	tarballPath, err := u.createTarball()
	if err != nil {
		return fmt.Errorf("tarball creation failed: %w", err)
	}

	// Upload tarball
	if err := u.uploadTarball(tarballPath); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	// Execute post-build commands
	if err := u.executePostBuildCommands(); err != nil {
		return fmt.Errorf("post-build commands failed: %w", err)
	}

	// Clean up local tarball
	if err := u.cleanupTarball(tarballPath); err != nil {
		u.logger.Warn("Failed to cleanup tarball: %v", err)
	}

	u.logger.Info("Uploader workflow completed successfully")
	return nil
}

func (u *Uploader) executePreBuildCommands() error {
	if len(u.config.PreBuildCommands) == 0 {
		return nil
	}

	u.logger.Info("Executing pre-build commands...")
	return utils.ExecuteCommands(u.config.PreBuildCommands, 5*time.Minute, u.logger)
}

func (u *Uploader) buildDockerImage() error {
	u.logger.Info("Building Docker image: %s:%s", u.config.ImageName, u.config.ImageTag)

	var buildCmd string
	if u.config.BuildCommand != "" {
		buildCmd = u.config.BuildCommand
	} else {
		buildCmd = fmt.Sprintf("docker build -t %s:%s %s",
			u.config.ImageName, u.config.ImageTag, u.config.DockerBuildPath)
	}

	output, err := utils.ExecuteCommand(buildCmd, 15*time.Minute)
	if err != nil {
		return err
	}

	u.logger.Debug("Docker build output: %s", strings.TrimSpace(output))
	return nil
}

func (u *Uploader) createTarball() (string, error) {
	u.logger.Info("Creating tarball for image: %s:%s", u.config.ImageName, u.config.ImageTag)

	// Generate tarball filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	tarballName := fmt.Sprintf("%s_%s_%s.tar", u.config.ImageName, u.config.ImageTag, timestamp)

	var tarballPath string
	if u.config.TarballPath != "" {
		if err := utils.EnsureDir(u.config.TarballPath); err != nil {
			return "", fmt.Errorf("failed to create tarball directory: %w", err)
		}
		tarballPath = filepath.Join(u.config.TarballPath, tarballName)
	} else {
		tarballPath = tarballName
	}

	// Save Docker image to tarball
	saveCmd := fmt.Sprintf("docker save %s:%s -o %s",
		u.config.ImageName, u.config.ImageTag, tarballPath)

	output, err := utils.ExecuteCommand(saveCmd, 10*time.Minute)
	if err != nil {
		return "", err
	}

	u.logger.Debug("Docker save output: %s", strings.TrimSpace(output))

	// Check if tarball was created successfully
	if !utils.FileExists(tarballPath) {
		return "", fmt.Errorf("tarball was not created: %s", tarballPath)
	}

	// Get and log tarball size
	size, err := utils.GetFileSize(tarballPath)
	if err != nil {
		u.logger.Warn("Failed to get tarball size: %v", err)
	} else {
		u.logger.Info("Tarball created: %s (%s)", tarballPath, utils.FormatBytes(size))
	}

	return tarballPath, nil
}

func (u *Uploader) uploadTarball(tarballPath string) error {
	u.logger.Info("Uploading tarball to %s@%s:%s", u.config.RemoteUser, u.config.RemoteHost, u.config.RemoteUploadPath)

	// Create SSH client
	client, err := u.createSSHClient()
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer client.Close()

	// Upload file using SCP
	if err := u.scpUpload(client, tarballPath); err != nil {
		return fmt.Errorf("SCP upload failed: %w", err)
	}

	u.logger.Info("Tarball uploaded successfully")
	return nil
}

func (u *Uploader) createSSHClient() (*ssh.Client, error) {
	// Read private key
	var auth []ssh.AuthMethod
	if u.config.RemoteKeyPath != "" {
		key, err := os.ReadFile(u.config.RemoteKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		auth = append(auth, ssh.PublicKeys(signer))
	}

	// Setup host key callback
	var hostKeyCallback ssh.HostKeyCallback
	knownHostsFile := filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")
	if utils.FileExists(knownHostsFile) {
		hkc, err := knownhosts.New(knownHostsFile)
		if err != nil {
			u.logger.Warn("Failed to load known_hosts, using insecure connection: %v", err)
			hostKeyCallback = ssh.InsecureIgnoreHostKey()
		} else {
			hostKeyCallback = hkc
		}
	} else {
		u.logger.Warn("known_hosts file not found, using insecure connection")
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	// Create SSH client config
	config := &ssh.ClientConfig{
		User:            u.config.RemoteUser,
		Auth:            auth,
		HostKeyCallback: hostKeyCallback,
		Timeout:         30 * time.Second,
	}

	// Connect to SSH server
	addr := fmt.Sprintf("%s:%d", u.config.RemoteHost, u.config.RemotePort)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	return client, nil
}

func (u *Uploader) scpUpload(client *ssh.Client, localPath string) error {
	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Get file info
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create remote file path
	fileName := filepath.Base(localPath)
	remotePath := filepath.Join(u.config.RemoteUploadPath, fileName)

	// Create SSH session
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Create SCP command
	scpCmd := fmt.Sprintf("scp -t %s", remotePath)

	// Get stdin pipe
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// Start SCP command
	if err := session.Start(scpCmd); err != nil {
		return fmt.Errorf("failed to start SCP command: %w", err)
	}

	// Send file header
	header := fmt.Sprintf("C%#o %d %s\n", fileInfo.Mode().Perm(), fileInfo.Size(), fileName)
	if _, err := stdin.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to send file header: %w", err)
	}

	// Copy file content
	if _, err := io.Copy(stdin, localFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Send end marker
	if _, err := stdin.Write([]byte("\x00")); err != nil {
		return fmt.Errorf("failed to send end marker: %w", err)
	}

	// Close stdin and wait for session to complete
	stdin.Close()
	if err := session.Wait(); err != nil {
		return fmt.Errorf("SCP command failed: %w", err)
	}

	return nil
}

func (u *Uploader) executePostBuildCommands() error {
	if len(u.config.PostBuildCommands) == 0 {
		return nil
	}

	u.logger.Info("Executing post-build commands...")
	return utils.ExecuteCommands(u.config.PostBuildCommands, 5*time.Minute, u.logger)
}

func (u *Uploader) cleanupTarball(tarballPath string) error {
	u.logger.Info("Cleaning up tarball: %s", tarballPath)
	return os.Remove(tarballPath)
}
