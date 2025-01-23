# Host Management

DevBox automatically manages your `/etc/hosts` file to provide local domain routing for your services.

Host management is configured in your `docker-compose.yml` using the `x-devbox-hosts` section.

## Example
```yaml
x-devbox-hosts:
  - ip: 127.0.0.1
    hosts:
      - "local.example.com"
      - "api.local.example.com"
      - "admin.local.example.com"
```

## Parameters

| Name | Required | Description |
| --- | --- | --- |
| ip | yes | The IP address to bind the hosts to |
| hosts | yes | The hostnames to add to the `/etc/hosts` file |


## Using Hosts

Your docker-compose manifest can contain multiple services with a gateway (like nginx or Traefik) to route requests. Configure your gateway to bind to 127.0.0.1 and route requests to your services.
