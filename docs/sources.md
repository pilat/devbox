# Sources Configuration

The `x-devbox-sources` block defines the repositories containing your service source code. Each repository will be cloned into the project's `sources` directory. DevBox automatically synchronizes all sources from their remote repositories when you run `devbox up`, `devbox update`, or `devbox restart`.

## Example
```yaml
x-devbox-sources:
  api:  # This name will be used as the directory name in ./sources
    url: https://github.com/company/monorepo.git
    branch: main 
    sparseCheckout:
      - backend/api

  dispatcher:
    url: https://github.com/company/monorepo.git
    branch: main
    sparseCheckout:
      - backend/dispatcher

  frontend:
    url: https://github.com/company/monorepo.git
```

## Parameters

| Name | Required | Description |
| --- | --- | --- |
| url | yes | Repository URL to clone (supports both HTTPS and SSH formats) |
| branch | no | Branch or tag to check out (defaults to repository's default branch) |
| sparseCheckout | no | List of specific paths to check out (useful for large repositories to reduce sync time) |

## Directory Structure

The example above creates the following directory structure:

```
.
└── .devbox/
    └── project-name/
        └── sources/
            ├── api/
            ├── dispatcher/
            └── frontend/
```

You can reference these source directories in your `docker-compose.yaml` file to ensure services use the latest source code:

```yaml
services:
  api:
    volumes:
      - ./sources/api:/app/api:ro
```

!!! tip "Mounting Source Code"
    You can mount source code without the `ro` (read-only) flag, but any files created in the `sources` directory will be lost during `devbox up`, `devbox update`, or `devbox restart` operations. For persistent changes, mount volumes below the source code directory.

## Local Development

Developers can use their own version of the source code instead of the managed one. See [Mount Sources](mount-sources.md) for details.

