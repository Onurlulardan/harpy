# Harpy

Harpy, adını harpi kartalından (harpy eagle) alır. Go ile yazılmış bir backend starter'dır: IAM, iş kuyruğu, worker, single-tenant ve multi-tenant, platform, composition, `core/cmd/api`, `core/cmd/worker`, `core/cmd/migrator`, `core/cmd/operator`, `cmd/setup`.

Yerel üçüncü parti Postgres 18, RabbitMQ ve RustFS'tir. Blob erişimi tek S3 API'dir (`aws-sdk-go-v2`).

Go module yolu `github.com/Onurlulardan/harpy`. Go sürümü `1.26`.

## Single-tenant

Klon: `git clone https://github.com/Onurlulardan/harpy.git`

İlk komut `go run ./cmd/setup`. Setup profili sorar, secret ve env yazar, sonraki komutları basar. Docker başlatmaz. Secret git'e girmez.

```text
go run ./cmd/setup
docker compose -f deploy/single-tenant/compose.yml up -d
go run ./core/cmd/migrator --env-file deploy/single-tenant/.env deploy up
go run ./core/cmd/operator --env-file deploy/single-tenant/.env initialize
go run ./core/cmd/api --env-file deploy/single-tenant/.env
go run ./core/cmd/worker --env-file deploy/single-tenant/.env
```

## Devam

1. [Harpy nedir](docs/harpy-starter/00-nedir.md)
2. [Mimari](docs/harpy-starter/01-mimari.md)
3. [Yeni modül](docs/harpy-starter/02-yeni-modul.md)
4. [fileguard ve archguard](docs/harpy-starter/03-fileguard-archguard.md)
