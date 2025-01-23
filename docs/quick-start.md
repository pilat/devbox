# Quick Start Guide

This guide will help you get started with DevBox by walking through the basic workflow.

## Initialize a Project

Initialize a new DevBox project from a manifest repository:

```bash
devbox init https://github.com/pilat/devbox-example
```

## Start Services

Start all services defined in your project (requires sudo password on first run):

```bash
devbox --name example-app up
```

DevBox will:

* Fetch source code from the manifest repository
* Pull required Docker images
* Build services as needed
* Configure SSL certificates
* Update host files
* Start all services

## Mount Source Code

To use your local source code instead of the manifest repository version:

```bash
cd /path/to/your/copy/of/service
devbox mount
```

## View Service Logs

```bash
# View logs for all services
devbox logs

# View logs for specific services
devbox logs service-name
```

## Stop Services

```bash
# Stop all services
devbox down
```
