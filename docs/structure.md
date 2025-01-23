# Project Structure

DevBox is designed to work with a dedicated repository that serves as a single source of truth for your local development environment. This repository is referred to as the **project repository** throughout the documentation.

!!! info "Repository Requirements"
    - A `docker-compose.yaml` file must be present in the repository root
    - The `sources` directory is not allowed
    - `.env` and `.devboxstate` files are not allowed

The repository can contain any additional files or directories needed, such as scripts, configuration files, or Dockerfiles.

## Project Initialization

When you run `devbox init <git-repo-url>` followed by `devbox up --name <project-name>`, DevBox creates the following structure on your machine:

```
~/.devbox/
├── project-name/           # Project name
|   |── .devboxstate        # DevBox's internal state
|   |── .env                # Environment variables managed by DevBox
|   |── .git                # Standard Git directory
│   ├── docker-compose.yml  # Manifest file from the project repository
│   └── sources/            # Sources managed by DevBox
│       ├── api/...
│       └── frontend/...
├── ca.pem                  # CA certificate managed by DevBox
└── ca.key                  # CA private key managed by DevBox
```

DevBox will synchronize this directory with the project repository whenever you run `devbox up`, `devbox update`, or `devbox restart`.

!!! warning "Important"
    The `~/.devbox/project-name` directory is managed by DevBox. Do not modify it manually. To make changes to the project, modify the remote repository and run `devbox update` to synchronize the changes.

## Extensions
DevBox extends the standard `docker-compose.yaml` file with additional configuration blocks using the `x-devbox-` prefix. For more information, see:

- [Sources](sources.md)
- [Certificates](certificates.md)
- [Hosts](hosts.md)
- [Scenarios](scenarios.md)
