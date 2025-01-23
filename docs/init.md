# Init Project

The `devbox init` command initializes a new DevBox project from a manifest repository.

## Usage

```bash
devbox init <git-source> [--name <project-name>] [--branch <branch-name>]
```

| Option | Required | Description |
| --- | --- | --- |
| `--name <project-name>` | no | Project name. If not specified, will be derived from the Git source |
| `--branch <branch-name>` | no | Branch name. If not specified, will use the repository's default branch |

## Example
```bash
devbox init https://github.com/pilat/devbox-example \
  --name example-app \
  --branch main
```

## Output

```
[*] Initializing project...

Project has been successfully initialized!
Next steps:
  1. Configure environment (optional):
     devbox --name example-app config env

  2. Start services:
     devbox --name example-app up
```
