# DevBox

DevBox is a lightweight CLI tool that streamlines local development for teams working on multi-service projects. It provides a way to use a single source of truth for local development environment configuration, eliminating common setup and maintenance challenges.

Built on top of `docker compose`, DevBox enhances the standard functionality with powerful features designed to improve team productivity.

## Introduction

Managing local development environments for projects with multiple interconnected services presents several challenges:
- Complex service configurations
- Version synchronization across teams
- SSL certificate management
- Host file configuration
- Source code management across repositories

DevBox solves these challenges by providing a centralized, automated approach to environment management, allowing teams to focus on development rather than infrastructure maintenance.

## Features

- **Automatic Environment Synchronization**  
  Keeps service configurations synchronized across the team. When configurations change or new services are added, everyone automatically gets the latest version.
  
- **Smart Source Code Management**  
  Automatically manages source code from multiple repositories, with support for:
  - Branch-specific checkouts
  - Sparse checkouts for monorepos
  - Easy switching between local and remote sources
  - Automatic synchronization with remote repositories

- **Development Flexibility**  
  Seamlessly switch between developing services locally and using their stable versions, perfect for working on specific components without managing the entire stack.

- **Intelligent Context Detection**  
  Automatically detects project and service context based on your current directory and Git information, reducing the need for explicit configuration.

- **Standardized Workflows**  
  Define common development tasks as scenarios in your configuration, making it easy to run tests, manage databases, or perform other routine operations consistently across the team.

- **Zero-Configuration SSL**  
  Automatically generates and manages SSL certificates for local development, with proper CA integration for both macOS and Linux.

- **Automated Host Management**  
  Manages your `/etc/hosts` file automatically, ensuring proper local domain routing without manual intervention.

## Installation

### Via Homebrew
```bash
brew tap pilat/devbox
brew install devbox
```

### From Binary Releases
1. Download the appropriate binary from the [Releases Page](https://github.com/pilat/devbox/releases)
2. Make the binary executable and add it to your PATH

### System Requirements
- Docker Engine 20.10.0 or later
- Git 2.28 or later

## Quick Start

### Initialize a Project
```bash
# Initialize from a manifest repository
devbox init https://github.com/pilat/devbox-example \
  --name example-app \
  --branch main
```

### Common Operations
```bash
# Start all services
devbox up

# View service status
devbox ps

# View service logs
devbox logs

# Use local source code
cd /path/to/your/service
devbox mount

# Revert to repository version
devbox unmount
```

## Documentation

For detailed documentation, see the [docs](docs) directory:
- [Project Structure](https://getdevbox.org/structure/)
- [Sources Configuration](https://getdevbox.org/sources/)
- [SSL Certificates](https://getdevbox.org/certificates/)
- [Host Management](https://getdevbox.org/hosts/)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

DevBox is licensed under the MIT License.  
© 2025 Vladimir Urushev
