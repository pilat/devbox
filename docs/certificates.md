# SSL Certificates

DevBox provides zero-configuration SSL certificate management for local development, automatically generating and managing certificates for your services.

SSL certificates are configured in your `docker-compose.yml` using the `x-devbox-cert` section.

## Example
```yaml
x-devbox-cert:
  domains: ["*.local.example.com", "local.example.com"]
  keyFile: "./certs/local.example.com.key.pem"
  certFile: "./certs/local.example.com.pem"
```

## Parameters

| Name | Required | Description |
| --- | --- | --- |
| domains | yes | The domains to generate certificates for |
| keyFile | yes | The path where the private key file will be stored |
| certFile | yes | The path where the certificate file will be stored |


## Using Certificates

Developers can use certificates by mounting them into services like nginx or Traefik.

!!! note "Creating Certificates"
    On first run, DevBox generates a root Certificate Authority (CA) pair (`~/.devbox/ca.crt` and `~/.devbox/ca.key`). To enable HTTPS for local development, DevBox automatically registers this CA with your system's trust store:

    - On macOS: Adds the CA to Keychain under the name "devbox development CA"
    - On Linux: Adds the CA to either `/usr/local/share/ca-certificates/devbox-ca.crt` or `/etc/pki/ca-trust/source/anchors/devbox-ca.crt` depending on your distribution

    DevBox then uses this CA to generate and sign certificates for your project's domains.

!!! warning ".gitignore"
    Certificate paths are relative to the project root and should be added to the project's `.gitignore` file. Otherwise, certificates will be deleted during project synchronization with the repository.
