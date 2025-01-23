# DevBox

DevBox is a specialized development environment orchestrator for complex, multi-service projects. Built on top of
Docker Compose, it manages source code, handles automatic updates, and simplifies service orchestration through a
single CLI.

---

## Introduction

Working on projects with multiple small, interconnected services can make setting up and maintaining a local development environment both challenging and time-consuming.
DevBox helps teams maintain a single source of truth for their local development environment, allowing everyone to focus on coding rather than troubleshooting and keeping the environment up to date.

---

## Key Features

- **Automatic Updates**  
  Ensures team-wide service configurations stay in sync. When a new service is added or a configuration changes, the latest version is available for everyone.
  
- **Centralized Repository Control**  
  Automatically fetches source code from stable branches of multiple repositories, reducing the need for manual version updates.

- **Selective Development**  
  Allows you to develop specific services locally while others remain at their default (latest) version.

- **Smart Context Detection**  
  Detects your location in the local filesystem to infer which project or service you’re working on.

- **Predefined Scenarios**  
  Provides scenarios for common tasks (e.g., running E2E tests) in a single configuration file, reducing the need to learn service-specific commands.

- **Zero-Config SSL**  
  Automatically generates and manages SSL certificates for local development.

- **Automatic Host File Updates**  
  Eliminates the need to manually edit `/etc/hosts` — DevBox takes care of it.

---

## Installation

DevBox is available via Homebrew, as binaries, or you can build it from source.

### Homebrew
```bash
brew tap pilat/devbox
brew install devbox
```

### Binaries
Download the latest binary from the [Releases Page](https://github.com/pilat/devbox/releases) and add it to your PATH.


## Quick Start

### Initialize a Project
```bash
# Initialize from a manifest repository
devbox init https://github.com/pilat/devbox-example \
  --name example-app \
  --branch main
```

### Basic Commands
```bash
# Start all services
devbox --name example-app up

# Mount local copy of the source code
cd /path/to/your/project
devbox mount

# Revert to the repository version of the source code
cd /path/to/your/project
devbox unmount
```

## Smart Context Detection
DevBox analyzes your current directory to determine the corresponding project and service:
- Project Detection: Matches your Git remote URLs or checks local mounts to identify the relevant DevBox project.
- Service Detection: Identifies which service you’re working on based on your directory structure and Git context.
With this feature, DevBox often eliminates the need to manually specify project or service names.


## Contributing
Contributions are always welcome! Please feel free to submit a Pull Request.


## License
DevBox is licensed under the MIT License.
© 2025 Vladimir Urushev
