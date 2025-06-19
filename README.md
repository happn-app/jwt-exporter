# JWT Exporter

A Prometheus exporter for JSON Web Tokens (JWTs)

## Configuration

See [example/config.yaml](https://github.com/happn-app/jwt-exporter/tree/master/example/config.yaml)

```yaml
address: :9100 # Address to listen to
label_selectors:
  - "monitor.jwt.io/monitoring=true"  # Labels to use to select secret for monitoring
polling_interval: 1h # Interval to poll kube API for secret changes
kubeconfig_path: ~/.kube/config # Path to kubeconfig, if an empty string, will use the in-cluster config
annotation_key: jwt-exporter/secret-key # Annotation that is used to find the key in the secret data containing the JWT
```

- Default config path: `/config/config.yaml`
- Default port: `:8080`
- Default metrics path: `/metrics`
- Default polling interval: `1h`

Docker image: `ghcr.io/happn-app/jwt-exporter:$VERSION`

CLI args:

- `-help`: show the help & usage
- `-config`: specify the config file path, supports `~` path expansion
- `-level`: specify the log level (default: info)

## Metrics exported

### Metrics

| **NAME**                                | **TYPE** | **DESCRIPTION**                             | **LABELS**                                                                                                    |
| --------------------------------------- | -------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `jwt_exporter_error_total`              | COUNTER  | JWT Exporter Errors                         | -                                                                                                             |
| `jwt_exporter_jwt_expires_in_seconds`   | GAUGE    | Number of seconds until the JWT expires.    | algorithm, audience, subject, id, scope, issuer, secret_key, secret_name, secret_namespace, name, email, role |
| `jwt_exporter_jwt_issued_since_seconds` | GAUGE    | Number of seconds since the JWT was issued. | algorithm, audience, subject, id, scope, issuer, secret_key, secret_name, secret_namespace, name, email, role |
| `jwt_exporter_jwt_expiration_timestamp` | GAUGE    | Timestamp of when the JWT expires.          | algorithm, audience, subject, id, scope, issuer, secret_key, secret_name, secret_namespace, name, email, role |
| `jwt_issued_at_timestamp`               | GAUGE    | Timestamp of when the JWT was issued.       | algorithm, audience, subject, id, scope, issuer, secret_key, secret_name, secret_namespace, name, email, role |

### Labels

| **LABEL**          | **EXAMPLE**                 | **DESCRIPTION**                                                                |
| ------------------ | --------------------------- | ------------------------------------------------------------------------------ |
| `id`               | -                           | ID of the JWT, corresponds to `jti` in the payload                             |
| `algorithm`        | `HS256`                     | Algorithm used in the JWT signature                                            |
| `email`            | `john.doe@example.com`      | Email the JWT was issued for                                                   |
| `name`             | `John Doe`                  | Name of the person the JWT was issued for                                      |
| `key_name`         | `jwt`                       | Name of the secret key the JWT was extracted from                              |
| `secret_name`      | `my-secret`                 | Metadata name of the secret the JWT was extracted from                         |
| `secret_namespace` | `default`                   | Namespace of the secret the JWT was extracted from                             |
| `audience`         | -                           | Audience of the JWT (`aud` in payload)                                         |
| `issuer`           | `https://auth.example.com/` | URL/ID of the issuer of the JWT                                                |
| `role`             | `reader`                    | One of the roles of the JWT, the metric is issued once per role and per scope  |
| `scope`            | `users:read`                | One of the scopes of the JWT, the metric is issued once per role and per scope |
| `subject`          | `61dc50123c1234567`         | The subject of the JWT, corresponds to the `sub` key in the payload            |

## RBAC

The exporter requires the ability to list secrets and namespaces, here is a sample `ClusterRole`:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: secret-reader
rules:
  - apiGroups: [""]
    resources: ["secrets", "namespaces"]
    verbs: ["get", "list"]

```
