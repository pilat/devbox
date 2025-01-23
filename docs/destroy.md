# Destroy Project

The `devbox destroy` command removes a project and all its resources completely.

## Usage

```bash
devbox destroy [--name <project-name>]
```

| Option | Required | Description |
| --- | --- | --- |
| `--name <project-name>` | no | Project name. If not specified, will be detected from Git source |

## Example
```bash
# Destroy current project
devbox destroy

# Destroy specific project
devbox --name project-name destroy
```

## Output

```bash
$ devbox destroy
[*] Down services...
 ⠿ Container api-1 Stopped
 ⠿ Container db-1 Stopped
 ⠿ Network example-app_default Removed

[*] Removing hosts...
[*] Removing project...

Project 'example-app' has been destroyed.
```
