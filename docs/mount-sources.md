# Mount Local Sources

The `devbox mount` command lets you use local source code instead of the stable sources in your project's containers.

--8<-- "auto-detect-note.md"

## Usage

```bash
devbox mount [--name <project-name>] [--source <source-name>] [--path <path-to-sources>]
```

| Option | Required | Description |
| --- | --- | --- |
| `--name <project-name>` | no | Project name. If not specified, will be detected from Git source |
| `--source <source-name>` | no | Source name. If not specified, will be detected from Git source |
| `--path <path-to-sources>` | no | Path to source code. If not specified, current directory will be used |

## Example
```bash
# Mount sources from current directory
cd /path/to/your/sources
devbox mount
```
