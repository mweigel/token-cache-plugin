# token-cache-plugin

Acquire and optionally cache bearer tokens for use with kubectl

# Purpose

Designed to work as a [credential plugin](https://kubernetes.io/docs/admin/authentication/#client-go-credential-plugins) to kubectl that
acquires a token from [kubernetes ldap](https://github.com/skippie81/kubernetes-ldap) and caches it locally so that users don't have to
enter their LDAP credentials every time kubectl is run. These tokens expire after a configurable amount of time which offers a balance
between security and ease of use.

## Configuration

Configuring a [credential plugin](https://kubernetes.io/docs/admin/authentication/#client-go-credential-plugins)

```yaml
apiVersion: v1
kind: Config
users:
- name: my-user
  user:
    exec:
      # Command to execute. Required.
      command: "token-cache-plugin"

      # API version to use when encoding and decoding the ExecCredentials
      # resource. Required.
      apiVersion: "client.authentication.k8s.io/v1alpha1"

      args:
      # Endpoint responsible for issuing tokens. Defaults to "".
      - '-token-request-endpoint=https://127.0.0.1:8443/ldapAuth'

      # Endpoint responsible for reviewing tokens. Defaults to "".
      - '-token-review-endpoint=https://127.0.0.1:8443/authenticate'

      # Path to CA certificate used to verify token request and token review endpoints. If not specified
      # the OS's default certificate store will be used.
      - '-ca-cert=/path/to/ca.pem'

      # Skip verification of the certificate presented by token request and token review endpoints.
      # Not recommended for producton environments. Defaults to false.
      - '-skip-tls-verification=true'

      # Whether to cache tokens returned by the token request endpoint. If tokens aren't cached then
      # credentials will have to be passed every time kubectl is run. This is meant to be used with
      # time restricted tokens. Derfaults to true.
      - '-cache-tokens=false'

      # Path to save locally cached tokens returned by the token request endpoint. Defaults to ~/.k8s-last-token
      - '-token-path='/fully/qualified/path/to/.token'
```

## Build

Dependencies managed by https://github.com/golang/dep

```bash
go build
```
