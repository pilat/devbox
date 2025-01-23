# Restart Services

The `devbox restart` command combines `devbox down` and `devbox up`. It stops and removes all services, then updates the project, fetches recent sources, and builds services before starting them.

--8<-- "auto-detect-note.md"

## Usage

```bash
devbox restart [--name <project-name>] [--profile <profile-name>]
```

| Option | Required | Description |
| --- | --- | --- |
| `--name <project-name>` | no | Project name. If not specified, will be detected from Git source |
| `--profile <profile-name>` | no | Profile to use from your `docker-compose.yml` file |

## Example
```bash
# Restart current project
devbox restart

# Restart specific project
devbox --name project-name restart
```

## Output

```
[*] Restart services...
[+] Running 3/3
 ✔ Container example-app-api-1         Restarted      1.2s
 ✔ Container example-app-frontend-1    Restarted      1.2s
 ✔ Container example-app-worker-1      Restarted      1.2s
```
