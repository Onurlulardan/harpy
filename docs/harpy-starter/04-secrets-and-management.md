# Secrets, rotation, and management bind

Setup contract for `go run ./cmd/setup`. The command does not start Docker.

## What a second setup keeps

| Material | Second setup | `--rotate-secrets` |
| --- | --- | --- |
| Postgres, RabbitMQ, RustFS, and initial admin passwords | kept | replaced |
| External S3 access key and secret supplied by the operator | kept | kept |
| Runtime cursor, IAM, FCART, notification, and actor-bridge key files | kept | replaced |
| Placeholder authority files such as `installation.json` | kept, including later edits | kept |
| Static local templates and password blocklist bytes | kept | kept |
| Complete `secrets/release/` bundle | kept byte for byte | rebuilt |
| Source-derived release manifests that do not contain the install key | kept | unchanged when the source corpus is unchanged |

Release private keys are not stored. A directory that is missing any catalogued release file is rebuilt as one set. Runtime files that are missing are created again; files that are still present are not replaced unless rotation was requested.

Deleting only one file of a derived pair, such as the actor-bridge seed without its public key, can leave that pair inconsistent. Remove the pair or run `--rotate-secrets`.

Refreshing local release evidence after a schema change requires `--rotate-secrets`. That rebuild also rotates runtime keys. Trust-root, signatures, receipts, and the catalog keyring change. Role, privilege, migration, extension, ownership, and operation documents stay put when their source corpus is unchanged.

Fingerprints in tests are SHA-256 of file bytes. Secret values are not written to stdout.

## Management listener

`/livez`, `/readyz`, and `/version` have no authentication.

| Endpoint | Response |
| --- | --- |
| `/livez` | Process is alive. Service name only. |
| `/readyz` | Dependency check names and ready or not-ready status. No credentials. |
| `/version` | Service, version, commit, build time, Go version, contract version, and environment. |

Local single-machine setup writes `HARPY_MANAGEMENT_LISTEN_ADDR=127.0.0.1:9090`. The public API stays on `127.0.0.1`. A second setup keeps a previously written address. `--management-listen host:port` replaces it.

Container deployment does not inherit the loopback default. `ops/local/compose.runtime.yml` sets `HARPY_MANAGEMENT_LISTEN_ADDR=0.0.0.0:9000` so the process accepts the published port. In-container health checks use `http://127.0.0.1:9000/readyz`. Host publish stays `127.0.0.1`. The process default used when the variable is absent remains `0.0.0.0:9090` in the runtime schema; local setup does not rely on that default.

## Fixture and generated hash commands

The recovery browser vector is the checked-in file `contracts/recoverykit/v1/testdata/browser-vector.json`. It is the canonical plaintext kit used by `ops/local/cmd/recovery-state`. It is not produced by a generator.

Currency registry provenance:

```text
go run ./tools/currencyregistry
go run ./tools/fileguard
```

The command records the SHA-256 of the existing generated Go and SQL files in `contracts/generated.yaml`. It does not rewrite those outputs. A second run on the same tree is byte-identical. `go run ./tools/archguard` is the import gate.
