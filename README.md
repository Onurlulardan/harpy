# Harpy

Harpy, adını harpi kartalından (harpy eagle) alır. Go ile yazılmış bir backend starter'dır: IAM, iş kuyruğu, worker, single-tenant ve multi-tenant, platform, composition, `core/cmd/api`, `core/cmd/worker`, `core/cmd/migrator`, `core/cmd/operator`, `cmd/setup`.

Yerel üçüncü parti Postgres 18, RabbitMQ ve RustFS'tir. Blob erişimi tek S3 API'dir (`aws-sdk-go-v2`).

Go module yolu `github.com/Onurlulardan/harpy`. Go sürümü `1.26`.

## Single-tenant

Klon: `git clone https://github.com/Onurlulardan/harpy.git`

Starter komut `task dev`. Sıra [AGENTS.md çalışma protokolü](AGENTS.md#çalışma-protokolü) kuralındadır.

## Devam

1. [Harpy nedir](docs/harpy-starter/00-nedir.md)
2. [Mimari](docs/harpy-starter/01-mimari.md)
3. [Yeni modül](docs/harpy-starter/02-yeni-modul.md)
4. [fileguard ve archguard](docs/harpy-starter/03-fileguard-archguard.md)
