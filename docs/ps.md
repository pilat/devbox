# Process Status

The `devbox ps` command displays real-time status information for all services in your project.

--8<-- "auto-detect-note.md"

## Usage

```bash
devbox ps [--name <project-name>]
```

| Option | Required | Description |
| --- | --- | --- |
| `--name <project-name>` | no | Project name. If not specified, will be detected from Git source |

## Example
```bash
# Show status of current project
devbox ps

# Show status of specific project
devbox --name project-name ps
```

## Output Format

Example output:
```
┌──────────┬─────────────┬─────────┬─────────┐
│ Age      │ Name        │ State   │ Health  │
├──────────┼─────────────┼─────────┼─────────┤
│ 2d       │ nginx       │ running │ healthy │
│ 01:23:45 │ api         │ running │ healthy │
│ 00:15:30 │ worker      │ running │ healthy │
│          │ db-migrate  │ exited  │         │
└──────────┴─────────────┴─────────┴─────────┘
```
