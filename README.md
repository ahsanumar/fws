# File Watch Server

A Golang application that automates Docker container deployment through file monitoring and SSH/SCP uploads. It operates in two modes:

1. **Uploader Mode**: Builds Docker images, creates tarballs, and uploads them to remote servers
2. **Watcher Mode**: Monitors directories for uploaded tarballs and automatically deploys them as containers

## Features

- 🐳 **Docker Integration**: Seamless Docker image building and container management
- 📦 **Tarball Creation**: Automatic Docker image export to tar archives
- 🚀 **SSH/SCP Upload**: Secure file transfer to remote servers
- 👁️ **File Monitoring**: Real-time file system watching with fsnotify
- 🔄 **Container Lifecycle**: Automatic container stop/start/replace
- 📊 **Logging**: Comprehensive logging with configurable levels
- 🎛️ **Configuration**: JSON-based configuration management
- 🔧 **CLI Interface**: User-friendly command-line interface

## Installation

### Prerequisites

- Go 1.21 or higher
- Docker installed and running
- SSH access to target servers (for uploader mode)

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd file-watch-server

# Build the application
make build

# Or install globally
make install
```

## Quick Start

### 1. Initialize Configuration

```bash
fws init
```

This creates a `config.json` file with default settings.

### 2. Configure the Application

Edit the `config.json` file to match your environment:

```json
{
  "mode": "watcher",
  "log_level": "info",
  "uploader": {
    "docker_build_path": "./",
    "image_name": "myapp",
    "image_tag": "latest",
    "tarball_path": "./tarballs",
    "remote_host": "destination.server.com",
    "remote_port": 22,
    "remote_user": "deploy",
    "remote_key_path": "~/.ssh/id_rsa",
    "remote_upload_path": "/opt/docker-uploads",
    "pre_build_commands": ["echo 'Starting build process...'"],
    "post_build_commands": ["echo 'Build process completed.'"]
  },
  "watcher": {
    "watch_directory": "/opt/docker-uploads",
    "container_name": "myapp",
    "container_ports": ["8080:8080"],
    "container_env": ["NODE_ENV=production"],
    "container_volumes": [],
    "pre_load_commands": ["echo 'Preparing to load new image...'"],
    "post_load_commands": ["echo 'New container deployed successfully.'"],
    "restart_policy": "unless-stopped"
  }
}
```

### 3. Run the Application

**Uploader Mode:**

```bash
fws --mode uploader --config config.json
```

**Watcher Mode:**

```bash
fws --mode watcher --config config.json --daemon
```

## Usage

### Command Line Options

```bash
Usage:
  fws [flags]
  fws [command]

Available Commands:
  init        Initialize configuration file
  status      Show container status (watcher mode only)
  logs        Show container logs (watcher mode only)
  help        Help about any command

Flags:
  -c, --config string   config file path (default: ./config.json)
  -d, --daemon          run as daemon in background
  -h, --help           help for fws
  -m, --mode string    operation mode: uploader or watcher
  -v, --verbose        verbose output (debug level)
```

### Uploader Mode

The uploader mode performs the following workflow:

1. **Pre-build Commands**: Execute custom commands before building
2. **Docker Build**: Build the Docker image from specified path
3. **Tarball Creation**: Export Docker image to tar archive
4. **SSH Upload**: Transfer tarball to remote server via SCP
5. **Post-build Commands**: Execute custom commands after upload
6. **Cleanup**: Remove local tarball file

Example uploader configuration:

```json
{
  "uploader": {
    "docker_build_path": "./app",
    "image_name": "mywebapp",
    "image_tag": "v1.0.0",
    "remote_host": "production.server.com",
    "remote_user": "deploy",
    "remote_key_path": "~/.ssh/deploy_key",
    "remote_upload_path": "/opt/deployments",
    "pre_build_commands": ["npm install", "npm run build"],
    "post_build_commands": ["echo 'Deployment package uploaded successfully'"]
  }
}
```

### Watcher Mode

The watcher mode monitors a directory and performs the following when a new tarball is detected:

1. **File Detection**: Monitor directory for `.tar` files
2. **Pre-load Commands**: Execute custom commands before processing
3. **Image Loading**: Load Docker image from tarball
4. **Container Management**: Stop and remove existing container
5. **Container Start**: Start new container with loaded image
6. **Post-load Commands**: Execute custom commands after deployment
7. **Cleanup**: Remove processed tarball

Example watcher configuration:

```json
{
  "watcher": {
    "watch_directory": "/opt/deployments",
    "container_name": "mywebapp",
    "container_ports": ["80:3000", "443:3443"],
    "container_env": [
      "NODE_ENV=production",
      "DATABASE_URL=postgresql://user:pass@localhost:5432/db"
    ],
    "container_volumes": ["/opt/app/data:/app/data", "/opt/app/logs:/app/logs"],
    "pre_load_commands": ["docker system prune -f"],
    "post_load_commands": [
      "curl -X POST http://localhost:3000/health",
      "echo 'Health check completed'"
    ],
    "restart_policy": "unless-stopped"
  }
}
```

## Configuration Reference

### Global Settings

- `mode`: Operation mode (`uploader` or `watcher`)
- `log_level`: Logging level (`debug`, `info`, `warn`, `error`)

### Uploader Configuration

- `docker_build_path`: Path to Dockerfile or build context
- `image_name`: Docker image name
- `image_tag`: Docker image tag
- `tarball_path`: Local directory to save tarballs
- `remote_host`: SSH server hostname/IP
- `remote_port`: SSH port (default: 22)
- `remote_user`: SSH username
- `remote_key_path`: Path to SSH private key
- `remote_upload_path`: Remote directory for uploads
- `build_command`: Custom Docker build command (optional)
- `pre_build_commands`: Commands to run before building
- `post_build_commands`: Commands to run after upload

### Watcher Configuration

- `watch_directory`: Directory to monitor for tarballs
- `container_name`: Name for the managed container
- `container_ports`: Port mappings (`["host:container"]`)
- `container_env`: Environment variables (`["KEY=value"]`)
- `container_volumes`: Volume mounts (`["host:container"]`)
- `pre_load_commands`: Commands before loading image
- `post_load_commands`: Commands after starting container
- `restart_policy`: Docker restart policy

## Monitoring and Management

### Check Container Status

```bash
fws status --config config.json
```

### View Container Logs

```bash
fws logs --config config.json
```

### Running as a System Service

Create a systemd service file:

```ini
[Unit]
Description=File Watch Server
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=deploy
WorkingDirectory=/opt/file-watch-server
ExecStart=/usr/local/bin/fws --config /opt/file-watch-server/config.json --daemon
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl enable file-watch-server
sudo systemctl start file-watch-server
```

## Security Considerations

1. **SSH Keys**: Use dedicated SSH keys with minimal permissions
2. **File Permissions**: Ensure proper file permissions on config files
3. **Network Security**: Configure firewall rules appropriately
4. **Container Security**: Use non-root users in Docker containers
5. **Directory Permissions**: Restrict access to upload directories

## Troubleshooting

### Common Issues

1. **Docker Permission Denied**: Add user to docker group

   ```bash
   sudo usermod -aG docker $USER
   ```

2. **SSH Connection Failed**: Verify SSH key permissions

   ```bash
   chmod 600 ~/.ssh/id_rsa
   ```

3. **File Watch Errors**: Check directory permissions and disk space

4. **Container Start Failures**: Verify Docker image integrity and port availability

### Debug Mode

Enable debug logging for detailed information:

```bash
fws --verbose --config config.json
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:

- Create an issue on the GitHub repository
- Check the troubleshooting section
- Review the configuration reference
