# Running Commands

The `devbox run` command executes predefined scenarios from your project configuration.

--8<-- "auto-detect-note.md"

## Usage

```bash
# Run a scenario
devbox run <scenario-name> [arg1 arg2 ...]

# Run a scenario in a specific project
devbox --name <project-name> run <scenario-name> [arg1 arg2 ...]
```

| Option | Required | Description |
| --- | --- | --- |
| `--name <project-name>` | no | Project name. If not specified, will be detected from Git source |

## Examples
```bash
# Run a test scenario
devbox run test

# Run a test scenario with arguments
devbox run test --verbose --suite=api

# Run a scenario in a specific project
devbox --name project-name run test
```

## See Also

- [Scenarios](scenarios.md)
