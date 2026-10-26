# Host KMS service

Owner: platform runtime tooling.

This command is the target-host TLS endpoint for the existing Google workload
identity and Cloud KMS clients. The explicit `local` profile supports disposable
development. The `host` profile validates API/worker Ed25519 assertions, issues
a five-minute bearer from a separate signing key, and authenticates every KMS
gRPC call.

The endpoint multiplexes these protocols on one listener:

- `POST /v1/token`: OIDC workload-identity token exchange response.
- Google Cloud KMS gRPC: `RawEncrypt` and `RawDecrypt` only.
- `GET /livez`: process health for the selected host profile.

Required environment:

- `HARPY_KMS_SERVICE_PROFILE`: exact `local` or `host`.
- `HARPY_KMS_SERVICE_LISTEN_ADDR`: TCP address, for example `:443`.
- `HARPY_KMS_SERVICE_TLS_CERT_FILE`: absolute PEM certificate path.
- `HARPY_KMS_SERVICE_TLS_KEY_FILE`: absolute PEM private-key path.
- `HARPY_KMS_SERVICE_MASTER_KEY_FILE`: absolute raw 32-byte AES key path.
- `HARPY_KMS_SERVICE_ACCESS_TOKEN_SIGNING_KEY_FILE`: host-only 32-byte key.
- `HARPY_KMS_SERVICE_WORKLOAD_PUBLIC_KEYS_FILE`: host-only workload registry.

The master and access-token keys are target-host generated state and must not
be committed. Verification covers config, authenticated STS, gRPC bearer
admission, resource-separated AES-GCM, CRC32C and TLS multiplexing.
