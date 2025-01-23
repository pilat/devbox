# Shell Access

The `devbox shell` command provides interactive shell access to running services. It automatically detects and uses the best available shell in the container, similar to `docker compose exec -it`.

--8<-- "auto-detect-note.md"

## Usage

```bash
devbox shell [--name <project-name>] <service-name>
```

| Option | Required | Description |
| --- | --- | --- |
| `--name <project-name>` | no | Project name. If not specified, will be detected from Git source |

## Example
```bash
# Access shell in a service
devbox shell service-name

# Access shell in a specific project's service
devbox --name project-name shell service-name
```

## Shell Detection

DevBox tries shells in the following order:
1. `/bin/zsh`
2. `/bin/bash`
3. `/bin/sh`
4. `/bin/ash`

The first available shell in the container will be used.
