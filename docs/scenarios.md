# Scenarios

Scenarios in DevBox are predefined commands that can be executed within your local services. They provide a convenient way to run common development tasks like running tests, managing databases, or starting development consoles.

Scenarios are configured in your `docker-compose.yml` using the `x-devbox-scenarios` section.

## Example
```yaml
x-devbox-scenarios:
  console:
    service: api
    description: "Run API console"
    command: ["bundle", "exec", "rails", "c"]
    interactive: true
    tty: true
    working_dir: /app
    user: app

  e2e:
    service: frontend
    description: "Run E2E tests"
    command: ["npm", "run", "test"]
```

## Parameters

| Name | Required | Description |
| --- | --- | --- |
| service | yes | The service to run the command in |
| description | no | A description of what the scenario does |
| command | yes | The command to run (as an array) |
| entrypoint | no | Override the container's entrypoint |
| tty | no | Whether to allocate a TTY (default: true) |
| interactive | no | Whether to run in interactive mode (default: true) |
| working_dir | no | Working directory inside the container |
| user | no | User to run as inside the container |


## Using Scenarios

Run scenarios using the `devbox run` command.

See [Run Scenarios](run.md) for details.
