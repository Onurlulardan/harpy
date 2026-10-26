# PostgreSQL 18.4 integration runtime

Phase 1 database, migration, privilege, cutover ve tenant-isolation testleri bu
ephemeral PG 18.4 + pgvector 0.8.3, UTF8/`en_US.utf8` ve data-checksum-on runtime'inda kosar. Production image family
multi-architecture OCI digest ile pinlidir;
mutable tag CI authority'si degildir. Database volume'u bilincli olarak
`tmpfs`tir, dolayisiyla her `down` sonrasi fresh-install testi yeniden baslar.

```powershell
docker compose -f ops/integration/compose.yml up -d --wait
$env:HARPY_TEST_POSTGRES_ADMIN_DSN = "postgres://harpy_bootstrap:harpy_integration_only@127.0.0.1:55432/harpy_integration?sslmode=disable"
go test -p=1 -timeout=45m -tags=integration ./...
docker compose -f ops/integration/compose.yml down
```

Port cakisirsa compose ve test ayni degeri kullanir:

```powershell
$env:HARPY_PG18_PORT = "55433"
```

Bu credential yalniz local/CI ephemeral cluster bootstrap'u icindir. Runtime,
migrator, adopter ve decoder rollerinin privilege negative testleri bootstrap
sonrasinda ayri login'lerle kosar; uygulama config'ine bu admin DSN verilmez.

`Dockerfile.pg-upgrade` ve `pg-upgrade.sh` hem representative test hem ADR-0037
full gate tarafindan kullanilir. Script `PG_UPGRADE_USER` ve
`POSTGRES_PASSWORD` olmadan baslamaz; pristine backup manifest'ini yeniden
dogrular, yalniz backup kopyasinda checksum acar ve optional evidence path'ine
strict `pg15-pg18-upgrade.v1` JSON kanitini no-replace yazar.
