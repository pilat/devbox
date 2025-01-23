# Starting Services

The `devbox up` command starts services in your DevBox project. It updates the project, fetches recent sources, and builds services before starting them, similar to `docker compose up -d`.

--8<-- "auto-detect-note.md"

## Usage

```bash
devbox up [--name <project-name>] [--profile <profile-name>]
```

| Option | Required | Description |
| --- | --- | --- |
| `--name <project-name>` | no | Project name. If not specified, will be detected from Git source |
| `--profile <profile-name>` | no | Profile to use from your `docker-compose.yml` file |

## Example
```bash
# Start current project
devbox up

# Start with specific profiles
devbox up --profile profile1 --profile profile2
```

## Output

```
[*] Updating project...
[*] Update hosts file...
[*] Setup CA...
[*] Generate certificates...
[*] Updating sources...
[+] Updating sources 4/4
 ✔ Source api-service      Synced                   2.1s
 ✔ Source web-frontend     Synced                   1.9s
 ✔ Source worker           Synced                   1.8s
 ✔ Source database         Synced                   1.7s

[*] Build services...
[+] Building 3/3
 ✔ Service api         Built                        1.4s
 ✔ Service frontend    Built                        2.5s
 ✔ Service worker      Built                        1.4s

[*] Up services...
[+] Running 3/3
 ✔ Network example-app_default         Created      0.0s
 ✔ Container example-app-api-1         Started      1.2s
 ✔ Container example-app-frontend-1    Started      1.2s
 ✔ Container example-app-worker-1      Started      1.2s
```

