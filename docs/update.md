# Update Project

The `devbox update` command synchronizes your project with its repository, updating host file entries and SSL certificates as needed.

--8<-- "auto-detect-note.md"

## Usage

```bash
devbox update [--name <project-name>]
```

| Option | Required | Description |
| --- | --- | --- |
| `--name <project-name>` | no | Project name. If not specified, will be detected from Git source |

## Example
```bash
# Update current project
devbox update

# Update specific project
devbox --name project-name update
```

## Output

```
[*] Updating project...
[*] Updating sources...
 ⠿ Source api-service: Synced
 ⠿ Source worker: Synced
 ⠿ Source frontend: Synced
 ⠿ Source shared-lib: Synced

 Sources:
┌────────────────┬────────────────────────┬──────────────┬─────────────┐
│ Name           │ Message                │ Author       │ Date        │
├────────────────┼────────────────────────┼──────────────┼─────────────┤
│ api-service    │ Update API endpoints   │ John Doe     │ 2 days ago  │
│ worker         │ Add job processor      │ Jane Smith   │ 1 hour ago  │
└────────────────┴────────────────────────┴──────────────┴─────────────┘
```
