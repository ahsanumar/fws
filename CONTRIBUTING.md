# Contributing to FWS (File Watch Server)

We welcome contributions to the File Watch Server project! This document provides guidelines for contributing.

## How to Contribute

### Reporting Issues

Before creating an issue, please check if it already exists. When creating an issue, include:

- **Clear Title**: Summarize the issue in the title
- **Description**: Provide a detailed description of the issue
- **Steps to Reproduce**: For bugs, include steps to reproduce the issue
- **Expected Behavior**: Describe what you expected to happen
- **Actual Behavior**: Describe what actually happened
- **Environment**: Include OS, version, and any relevant configuration
- **Logs**: Include relevant log output if applicable

### Feature Requests

For feature requests, please include:

- **Use Case**: Describe why this feature would be useful
- **Proposed Implementation**: If you have ideas on how to implement it
- **Alternatives**: Any alternative solutions you've considered

### Pull Requests

1. **Fork the Repository**: Create a fork of the repository on GitHub

2. **Create a Branch**: Create a feature branch from `main`

   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make Changes**:

   - Write clear, concise commit messages
   - Follow the existing code style
   - Add tests for new functionality
   - Update documentation as needed

4. **Test Your Changes**:

   ```bash
   # Run tests
   go test ./...

   # Run linter
   make lint

   # Build and test
   make build
   ./fws --help
   ```

5. **Update Documentation**:

   - Update README.md if needed
   - Update CHANGELOG.md
   - Add/update code comments

6. **Submit Pull Request**:
   - Create a pull request against the `main` branch
   - Include a clear description of the changes
   - Reference any related issues

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Docker (for testing)
- Make
- Git

### Local Development

1. **Clone the repository**:

   ```bash
   git clone https://github.com/ahsanumar/fws.git
   cd fws
   ```

2. **Install dependencies**:

   ```bash
   go mod download
   ```

3. **Build the project**:

   ```bash
   make build
   ```

4. **Run tests**:

   ```bash
   make test
   ```

5. **Run linting**:
   ```bash
   make lint
   ```

### Code Style Guidelines

- **Go Format**: Use `gofmt` to format your code
- **Naming**: Follow Go naming conventions
- **Comments**: Add comments for public functions and complex logic
- **Error Handling**: Always handle errors appropriately
- **Logging**: Use the internal logger with appropriate levels

### Testing

- Write unit tests for new functionality
- Test edge cases and error conditions
- Use table-driven tests where appropriate
- Ensure tests are deterministic and don't depend on external services

### Commit Messages

Use clear, descriptive commit messages:

```
feat: add support for custom Docker build args
fix: resolve race condition in file watcher
docs: update installation instructions
test: add unit tests for config validation
```

Prefix types:

- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation changes
- `test`: Test additions/modifications
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

## Project Structure

```
fws/
├── cmd/                    # CLI commands
├── internal/
│   ├── config/            # Configuration management
│   ├── uploader/          # Upload functionality
│   ├── watcher/           # File watching functionality
│   └── utils/             # Utility functions
├── examples/              # Example configurations and demos
├── .github/               # GitHub workflows and templates
├── docs/                  # Documentation
└── scripts/               # Build and utility scripts
```

## Release Process

1. Update CHANGELOG.md with new changes
2. Create a new tag following semantic versioning
3. Push the tag to trigger the release workflow
4. GitHub Actions will automatically build and create the release

## Questions?

If you have questions about contributing, feel free to:

- Open an issue for discussion
- Contact the maintainers
- Check existing issues and pull requests

Thank you for contributing to FWS! 🚀
