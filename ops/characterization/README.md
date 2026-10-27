# Izole V1 Oracle Runtime

Bu dizin V1 davranisini kaydetmek ve replay etmek icin tek kalici runtime
yuzeyidir. `backend/docker-compose.yml` dahil edilmez veya override edilmez.
Stack host portu, sabit `container_name`, external network ya da kalici runtime
volume'u tanimlamaz.

## Guvenlik siniri

- Normal record/replay PostgreSQL, Redis, RabbitMQ, MinIO ve auth backup state'i
  `tmpfs` uzerindedir. PG18 gate kendi project-scoped named volume'larini kullanir.
- Migration, seed ve backfill DB host'u Compose icindeki `postgres` adina
  hard-code edilmistir; `.env` ile developer DB'ye yoneltilemez.
- Oracle runtime image'i `reset-db` ve `adjust-db` binary'lerini tasimaz.
- Disari yalniz redacted artifact ve atomik uretilen oracle license-cache bind
  dizinleri yazilir.
- Runtime network `internal: true`dir; SMTP ve AI provider'a cikis
  yoktur.
- Her calisma `harpy-v1-oracle` project-name prefix'iyle baslamak zorundadir.

Bu sinirlar nedeniyle bu stack mevcut developer container, volume, network veya
database'ine baglanamaz. `cmd/reset-db` bu lifecycle'in parcasi degildir.

## Girdiler

`.env.example` yalniz degisken adlarini tasir. `.env` commit edilmez. Image
degerleri tam `name:tag@sha256:digest` olmali; PostgreSQL PG15, V1 build Go
1.25 ve backend-v2 characterizer build Go 1.26.5 baseline'ini gostermelidir.
`V1_ORACLE_SOURCE_VERSION`, build edilen V1 source snapshot'inin lowercase
hash'idir.

`V1_ORACLE_LICENSE_CACHE_DIR` mevcut ve container UID `65532` icin yazilabilir
bir dizindir. `license-cache-init`, canonical
`tools/v1characterize license-cache` komutuyla V1 source contract hash'lerini
attest eder, explicit
oracle girdilerinden `state.json` dosyasini atomik uretir ve ayni seal ile tekrar
dogrular. Herhangi bir adim basarisizsa Gateway baslamaz. Gateway ayni dizini
read-only mount eder.

Seal, `.env` icindeki `HARPY_LICENSE_KEY` ile V1'in gercek instance-ID ve
cache-key algoritmasindan turetilir. Secret flag, command argument, stdout veya
fixture alani olmaz. Uretici yalniz network'suz oracle init container'inda
bulunur; production Gateway image'ine ya da runtime flag'ine eklenmez.

Artifact host dizini calismadan once mevcut olmalidir. Suite kontratlari
characterizer image'ina source ile birlikte gomulur; V1 source ise ayni build
allowlist'ini attest etmek icin read-only mount edilir. Compose
`create_host_path: false` kullandigi icin yazim hatali veya eksik path sessizce
dizin olusturmaz.

Runner image yalniz canonical `tools/v1characterize` komutunu source'tan build
eder. Komut yoksa build fail eder; shell fallback veya ikinci runner yoktur.

## Zorunlu lifecycle

Komutlar bu dizinde calistirilir:

```powershell
Copy-Item .env.example .env
New-Item -ItemType Directory -Force .license-cache, .artifacts
```

`.env` doldurulur; license cache init container tarafindan uretilir. Ardindan:

```powershell
docker compose --env-file .env -f compose.yml config --quiet
docker compose --env-file .env -f compose.yml --profile characterize build api-gateway auth-service rabbitmq license-cache-init characterizer
docker compose --env-file .env -f compose.yml up --wait
docker compose --env-file .env -f compose.yml run --rm --no-deps characterizer
```

`up` sirasini Compose kapatir:

```text
preflight -> postgres -> migrate -> seed -> record-index backfill -> V1 services
preflight -> source-attested license-cache-init ----------------> Gateway
```

Seed yalniz `oracle_a` tenant'ini kurar. `oracle_b`, operator ve portal actor
yalniz characterizer tarafindan gercek Gateway/API akislariyla olusturulur;
urun actor'u direct SQL ile eklenmez. Recorder Gateway ile ayni network
namespace'ini paylasir ve urun trafigini exact loopback `127.0.0.1:8000`
uzerinden gonderir. Remote host, servis DNS adi ve redirect kabul edilmez;
direct PostgreSQL ve RabbitMQ erisimi yalniz gozlemdir.

Characterizer runtime servislerine `depends_on` tanimlamaz. Servis sagligi
yalniz onceki `up --wait` kapisiyla kurulur ve `run --no-deps`, migration/seed
init container'larinin assertion sirasinda yeniden calismasini engeller.

Her basarili veya basarisiz calisma sonunda:

```powershell
docker compose --env-file .env -f compose.yml down --volumes --remove-orphans
```

`down --volumes` savunma amaclidir; runtime state zaten `tmpfs`tir. Project-name
prefix'i disindaki bir Compose projesine teardown uygulanmaz.

## PostgreSQL 18 certification lifecycle

Authoritative operator komutu ADR-0037'deki yuzeydir:

```powershell
go run ./tools/v1characterize pg18-gate -oracle-env ops/characterization/.env
```

Release pipeline su path-only degerleri process environment'ina verir:
`HARPY_RELEASE_MANIFEST_PATH`, `HARPY_RELEASE_SIGNATURE_PATH`,
`HARPY_RELEASE_TRUST_ROOT_PATH`, `HARPY_ATTESTATION_SIGNING_KEY_FILE`,
`HARPY_RELEASE_DATABASE_NAME`,
`HARPY_PG18_CERTIFICATION_ARTIFACT_DIR` ve
`HARPY_PG18_CERTIFICATION_RECEIPT_DIR`. Bunlar ayni adli CLI flag'leriyle
explicit override edilebilir; private key byte'i flag veya Docker environment'i
olmaz. `-promoted-receipt` yalniz explicit immutable promotion icindir.

Host her kosu icin run-ID'li candidate, baseline ve receipt adi uretir; paralel
kosular ayni artifact'a yazamaz. Candidate dosyalari hostta tekrar strict load
ve hash edilir. Cleanup basarmadan receipt imzalanip publish edilmez; publish
sonrasi receipt diskten tekrar okunup imza/release/policy bagiyla dogrulanir.

Komut her kosuda unique `harpy-v1-pg18-gate-*` project adi uretir. Mutable
tag, baska project prefix'i veya Docker-unavailable durumu skip edilmez.
`compose.pg18-source.yml` bos `pg15-source` volume'unda migrate, seed ve
record-index backfill'i bir kez kosar. Salt-okunur `database-ready` sonrasinda
`pg15-backup` fiziksel yedegi `pg_verifybackup` ile dogrular ve su ara kanitlari
`pg18-evidence` volume'una no-replace yazar:

```text
/evidence/backup.json
/evidence/backup-manifest.sha256
```

`pg18-upgrade` yalniz backup volume'unu recovery ile clean kapatir, offline
checksum enable/check yapar ve exact PG18 target volume'unu `pg_upgrade
--check` + `--copy` ile uretir. Upgrade kullanicisi required
`PG_UPGRADE_USER=oracle` kontratidir. Sonuc `/evidence/upgrade.json` olur.

Source container'lari volume silmeden kapatilir. `compose.pg18-target.yml` bos
volume'u init etmez; `PG_VERSION=18` yoksa baslamaz. DDL-free `database-ready`
source system identifier'inin degismesini ve database OID'nin korunmasini exact
dogrular. Ardindan `pg18-analyze`, fresh PgBouncer/Redis/RabbitMQ/MinIO ve tum V1
servisleri baslar. Target compose migrate, seed veya backfill servisi icermez.

Tum cikislarda yalniz ayni project adiyla su cleanup yuzeyi kullanilir:

```powershell
docker compose --project-name $project --env-file .env `
  -f compose.pg18-cleanup.yml down --volumes --remove-orphans
```

Cleanup compose dort named volume'u explicit mount ederek source, backup,
target ve evidence state'inin birlikte silinmesini garanti eder. Global prune,
label-disindaki container silme veya production source volume cleanup'i yoktur.

## Moduler dosyalar

| Dosya | Owner |
| --- | --- |
| `compose.infra.yml` | PG15, PgBouncer, Redis, RabbitMQ, MinIO |
| `compose.preflight.yml`, `compose.license.yml` | preflight ve deterministic license cache |
| `compose.init.yml` | migrate/seed/backfill ve source readiness |
| `compose.services-platform.yml` | auth, permission, notification, document, log |
| `compose.services-engines.yml` | Forms, Workflow, OpenAPI, AI Chat, MCP |
| `compose.gateway.yml` | public V1 Gateway ve license mount |
| `compose.runner.yml` | ayni networkte source-built characterizer |
| `compose.pg18-source.yml` | named PG15 init, backup ve physical upgrade |
| `compose.pg18-target.yml` | DDL-free PG18 readiness ve full V1 runtime |
| `compose.pg18-cleanup.yml` | project-scoped container/network/volume cleanup |
