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
      #
      # The API version returned by the plugin MUST match the version encoded.
      apiVersion: "client.authentication.k8s.io/v1alpha1"

      # Environment variables to set when executing the plugin. Optional.
      env:
        # URL of service responsible for issuing bearer tokens. Required.
      - name: TOKEN_SERVER_URL
        value: "https://127.0.0.1:443/ldapAuth"

        # Path to CA certificate used to verify token server's certificate.
        # If not specified default certificate store will be used. Optional.
      - name: CA_CERT
        value: "/path/to/ca-cert"

        # Whether to cache tokens locally for reuse. Default true. Optional.
      - name: CACHE_TOKENS
        value: "true"

```
