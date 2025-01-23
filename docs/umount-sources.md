# Unmount Local Sources

The `devbox umount` command removes local source code mounts from your project's containers, reverting to the stable sources.

--8<-- "auto-detect-note.md"

## Usage

```bash
devbox umount [--name <project-name>] [--source <source-name>]
```

| Option | Required | Description |
| --- | --- | --- |
| `--name <project-name>` | no | Project name. If not specified, will be detected from Git source |
| `--source <source-name>` | no | Source name. If not specified, will be detected from Git source |

## Example
```bash
# Unmount sources from current directory
devbox umount
```
