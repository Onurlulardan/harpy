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

The actor-bridge seed and its public key file are one derived pair. If either file is missing, the next setup rewrites both together. Other runtime files are still recreated individually.

`--refresh-release` rebuilds `secrets/release/` without rotating runtime keys or compose passwords. The release private key is not stored, so this rebuild mints a new release trust root. Role, privilege, migration, extension, ownership, and operation documents stay put when their source corpus is unchanged. `--rotate-secrets` still rotates runtime keys and the release trust material together.

Fingerprints in tests are SHA-256 of file bytes. Secret values are not written to stdout.

## Restore a copied secrets tree

Copy `deploy/<profile>/secrets` back into place, then run setup again. SHA-256 fingerprints stay. A normal second setup does not replace the cursor keyring, the IAM keyrings, the FCART KEK keyrings, or the release trust root. `runtime/access-token-keyring.json` still parses after that second setup.

`--rotate-secrets` invalidates the previous cursor keyring, IAM keyrings, FCART KEK keyrings, and release trust root. The table above still applies: compose passwords and the other listed runtime key files are replaced, and operator-supplied external S3 keys stay.

## Preserved authority and cutover windows

`--rotate-secrets` keeps the installed issue clock on `authority-manifest.json` and `cutover-operation.json`. The new trust root still changes. When the preserved `issued_at` is earlier than the new release clock, the new trust root `not_before` reaches back to one hour before that `issued_at`, so the new key can sign the unchanged documents. The cutover receipt `adopted_at` stays inside the original operation window, including the 12-minute cutover window. A later rotate leaves those original validity windows in place.

## What is deterministic

Source-derived and deterministic: the currency Go file, a fresh currency SQL seed, a fresh migration manifest, and the release index. Per-install and not deterministic: the runtime keyring, compose passwords, and the release trust root. Hashes in `contracts/generated.yaml` are not hand-edited.

## Management listener

`GET /livez`, `/readyz`, and `/version` return 200 with no auth header. The JSON has no password, token, DSN, or key material.

| Endpoint | Response |
| --- | --- |
| `/livez` | Process is alive. Service name only. |
| `/readyz` | Dependency check names and ready or not-ready status. No credentials. |
| `/version` | Service, version, commit, build time, Go version, contract version, and environment. |

| Placement | Who can connect |
| --- | --- |
| Local setup | Same machine on loopback `127.0.0.1:9090`. Setup writes `HARPY_MANAGEMENT_LISTEN_ADDR=127.0.0.1:9090`. A later setup keeps it unless `--management-listen` overrides it. The public API stays on `127.0.0.1`. |
| Container process | Listens on `0.0.0.0:9000` so the process accepts the published port. In-container health is `http://127.0.0.1:9000/readyz`. |
| Host publish | `127.0.0.1` only. |

Residual risk: a hand-written env that omits `HARPY_MANAGEMENT_LISTEN_ADDR` still binds `0.0.0.0:9090`. That is the schema default in `core/internal/platform/config/runtime.go`. Local setup and compose always write an address, so they do not use that default.

## Fixture and generated hash commands

The recovery browser vector is the checked-in file `contracts/recoverykit/v1/testdata/browser-vector.json`. It is the canonical plaintext kit used by `ops/local/cmd/recovery-state`. It is not produced by a generator.

Currency registry provenance:

```text
go run ./tools/currencyregistry
go run ./tools/fileguard
```

The command records the SHA-256 of the existing generated Go and SQL files in `contracts/generated.yaml`. It does not rewrite those outputs. A second run on the same tree is byte-identical and writes nothing when the hashes match. The writer splices only the two currency records. CI runs the drift gate in [03-fileguard-archguard.md](03-fileguard-archguard.md).
