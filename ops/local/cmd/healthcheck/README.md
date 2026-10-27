# Local healthcheck

Owner: platform local-development tooling.

This local-only command is shared by container healthchecks and Task verification.
It retries every configured target until one probe round reports all targets
healthy or the total timeout expires.

```text
healthcheck \
  -http https://service:9000/readyz \
  -http http://service:8000/livez \
  -tcp postgres:5432 \
  -ca-file /run/local/ca.crt \
  -timeout 30s
```

- `-http` and `-tcp` are repeatable; at least one target is required.
- HTTP accepts only `http` or `https` URLs without credentials, query, or fragment.
- HTTP readiness requires a `2xx` response; redirects are not followed.
- `-ca-file` adds a PEM CA to the system trust roots. TLS verification is never
  disabled.
- `-timeout` is the bounded total retry budget.

The helper has no product, tenant, permission, persistence, event, job, V1, or
public API impact. Rollback is removal from local image and Compose wiring.
