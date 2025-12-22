# Viewing Logs

The `devbox logs` command displays and follows logs from your services, similar to `docker compose logs --tail 500 --follow`.

--8<-- "auto-detect-note.md"

## Usage

```bash
devbox logs [--name <project-name>] [<service-name-1> <service-name-2> ...]
```

| Option | Required | Description |
| --- | --- | --- |
| `--name <project-name>` | no | Project name. If not specified, will be detected from Git source |
| `<service-name-1> <service-name-2> ...` | no | Service names to show logs for. If not specified, shows logs from all services |

## Example
```bash
# View logs from all services
devbox logs

# View logs from specific services
devbox logs service1 service2

# View logs from a specific project
devbox --name project-name logs
```

## Log Size Limits

DevBox automatically limits container logs to **10MB per service** to prevent disk space exhaustion. This is applied to all services that don't have explicit logging configuration.

To override for a specific service, add a `logging` section in your `docker-compose.yaml`:

```yaml
services:
  verbose-service:
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
```
