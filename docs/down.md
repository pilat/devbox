# Stopping Services

The `devbox down` command stops and removes all services in your project, similar to `docker compose down`.

--8<-- "auto-detect-note.md"

## Usage

```bash
devbox down [--name <project-name>]
```

| Option | Required | Description |
| --- | --- | --- |
| `--name <project-name>` | no | Project name. If not specified, will be detected from Git source |

## Example
```bash
# Stop current project
devbox down

# Stop specific project
devbox --name project-name down
```

## Output

```
[*] Down services...
[+] Running 3/3
 ✔ Container example-app-api-1         Removed      1.2s
 ✔ Container example-app-frontend-1    Removed      1.2s
 ✔ Network example-app_default         Removed      0.0s
```
