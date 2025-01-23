# Environment Configuration

The `devbox config env` command configures environment variables for your project.

--8<-- "auto-detect-note.md"

## Usage

```bash
# Configure current project
devbox config env

# Configure specific project
devbox --name project-name config env
```

| Option | Required | Description |
| --- | --- | --- |
| `--name <project-name>` | no | Project name. If not specified, will be detected from Git source |

## Result

The command opens your default text editor (usually `vi` or `vim`) to edit the project-specific `.env` file. Make your changes and save the file.

## Next Steps
- [Update project](update.md)
