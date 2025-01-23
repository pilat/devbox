# Project Info

The `devbox info` command fetches the latest manifest, updates host file entries, updates SSL certificates, and displays detailed information about your project's sources and mounted volumes.

--8<-- "auto-detect-note.md"

## Usage

```bash
devbox info [--name <project-name>]
```

| Option | Required | Description |
| --- | --- | --- |
| `--name <project-name>` | no | Project name. If not specified, will be detected from Git source |

## Example
```bash
# Get info for current project
devbox info

# Get info for specific project
devbox --name example-app info
```

## Output

```
 Sources:
┌────────────────┬────────────────────────┬──────────────┬─────────────┐
│ Name           │ Message                │ Author       │ Date        │
├────────────────┼────────────────────────┼──────────────┼─────────────┤
│ api-service    │ Update API endpoints   │ John Doe     │ 2 days ago  │
│ worker (jobs)  │ Add job processor      │ Jane Smith   │ 1 hour ago  │
└────────────────┴────────────────────────┴──────────────┴─────────────┘

 Mounts:
┌────────────────────┬───────────────────────────────┐
│ Mount path         │ Local path                    │
├────────────────────┼───────────────────────────────┤
│ ./sources/api      │ /Users/dev/code/api-service   │
└────────────────────┴───────────────────────────────┘
```
