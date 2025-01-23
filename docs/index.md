# What is DevBox?

DevBox is a lightweight CLI tool that streamlines local development for teams working on multi-service projects. It provides a way to use a single source of truth for local development environment configuration, eliminating common setup and maintenance challenges.

Built on top of `docker compose`, DevBox enhances the standard functionality with powerful features designed to improve team productivity.


## Introduction

Setting up and maintaining a local development environment for projects with multiple interconnected services can be challenging and time-consuming. DevBox helps teams maintain a single source of truth for their local development environment configuration, allowing everyone to focus on coding rather than troubleshooting environment issues.


## Features

### Automatic Updates
Ensures team-wide service configurations stay in sync. When services are added or configurations change, the latest version is immediately available for everyone.

### Centralized Repository Control
Automatically fetches source code from stable branches of multiple repositories, eliminating manual version updates.

### Selective Development
Develop specific services locally while others remain at their default (latest) version.

### Smart Context Detection
Automatically detects your location in the local filesystem to infer which project or service you're working on:

* Project Detection: Identifies the relevant DevBox project by matching Git remote URLs or checking local mounts
* Service Detection: Determines which service you're working on based on directory structure and Git context

### Predefined Scenarios
Provides scenarios for common tasks (e.g., running E2E tests) in a single configuration file, eliminating the need to learn service-specific commands.

### Zero-Config SSL
Automatically generates and manages SSL certificates for local development.

### Automatic Host File Updates
Manages your `/etc/hosts` file automatically — no manual editing required.
