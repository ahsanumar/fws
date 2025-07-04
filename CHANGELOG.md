# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial public release preparation
- GitHub Actions CI/CD workflows
- Automated release generation
- Installation script for easy deployment

## [v1.0.0] - 2024-07-04

### Added

- File watching functionality with fsnotify
- Docker image building and tarball creation
- SSH/SCP upload capabilities
- Container lifecycle management (stop/remove/start)
- Dual mode operation (uploader/watcher)
- JSON-based configuration system
- CLI interface with Cobra
- Daemon mode support with signal handling
- Comprehensive logging with configurable levels
- Pre/post command hooks for custom workflows
- Container status and log viewing
- Cross-platform build support (Linux, macOS, ARM64)
- Example configurations and sample application
- Comprehensive documentation and README
- Makefile for build automation
- Quick start script for demonstration

### Features

- **Uploader Mode**:

  - Builds Docker images from source
  - Creates tarballs from Docker images
  - Uploads via SSH/SCP to remote servers
  - Supports custom build commands
  - Pre/post-build command execution

- **Watcher Mode**:

  - Monitors directories for new tarballs
  - Automatically loads Docker images
  - Manages container lifecycle
  - Supports custom port mappings and environment variables
  - Pre/post-load command execution
  - Configurable restart policies

- **Configuration**:

  - JSON-based configuration management
  - Validation and error handling
  - Support for SSH key authentication
  - Flexible directory and path configuration

- **CLI & Management**:

  - User-friendly command-line interface
  - Container status monitoring
  - Log viewing capabilities
  - Init command for quick setup
  - Verbose logging option

- **Development & Deployment**:
  - Cross-platform compilation
  - Systemd service support
  - Docker containerization support
  - Comprehensive build automation

### Security

- SSH key-based authentication
- Known hosts verification with fallback
- Secure file transfer protocols
- Proper error handling and logging

### Documentation

- Comprehensive README with examples
- Configuration reference
- Troubleshooting guide
- Quick start tutorial
- Security considerations
- Installation and deployment instructions

[Unreleased]: https://github.com/ahsanumar/fws/compare/v1.0.0...HEAD
[v1.0.0]: https://github.com/ahsanumar/fws/releases/tag/v1.0.0
