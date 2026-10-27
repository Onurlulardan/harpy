# Harpy

Harpy, adını harpi kartalından (harpy eagle) alır. Go ile yazılmış bir backend starter'dır: IAM, iş kuyruğu, worker, single-tenant ve multi-tenant, platform, composition ve `core/cmd/api`, `core/cmd/worker`, `core/cmd/migrator`, `core/cmd/operator`, `cmd/setup`. Yerel üçüncü parti Postgres 18, RabbitMQ ve RustFS'tir. Blob erişimi tek S3 API'dir (`aws-sdk-go-v2`).

`forms`, `workflow`, AI modülü, MCP sunucusu ve tarihsel SQL migration dökümleri starter'da yoktur. Migrator kodu starter'dadır. Şema SQL'i Harpy için yazılır; eski bir SQL dökümü olarak alınmaz. `work` vardır; workflow motoru değildir.

Go module yolu `github.com/Onurlulardan/harpy`. Go sürümü `1.26`.

## Klon

```text
git clone --branch monorepo https://github.com/Onurlulardan/harpy.git
cd harpy
```

## İlk komut

İlk komut:

```text
go run ./cmd/setup
```

Setup profili sorar, secret ve env yazar, sonraki komutları basar. Docker başlatmaz. Secret git'e girmez. `--defaults` soru sormaz.

Profil `single-tenant` veya `multi-tenant` olur.

## Setup'ın bastığı komutlar

```text
docker compose -f deploy/<profil>/compose.yml up -d
go run ./core/cmd/migrator --env-file deploy/<profil>/.env deploy up
go run ./core/cmd/operator --env-file deploy/<profil>/.env initialize
go run ./core/cmd/api --env-file deploy/<profil>/.env
go run ./core/cmd/worker --env-file deploy/<profil>/.env
```

`<profil>`, seçilen `single-tenant` veya `multi-tenant` dizinidir. `single-tenant`, `DEPLOYMENT_MODE=onpremise` (müşteri tenant tavanı 0) yazar. `multi-tenant`, `DEPLOYMENT_MODE=saas` yazar.

Modül yolunu değiştirmek için `go.mod` içindeki `module` satırını güncelleyin, sonra `go run ./tools/archguard` çalıştırın.

Sıra: klon, setup, compose, migrate, operator initialize, api, worker.

Ayrıntı: [İlk çalıştırma, tenant profilleri, RustFS](docs/harpy-starter/03-ilk-calistirma-tenant-rustfs.md).

## Obje depolama

Yerel obje servisi RustFS'tir. Uygulama S3 konuşur (`aws-sdk-go-v2`). Amazon S3 ve MinIO aynı kapının başka endpoint'idir. MinIO imajı yoktur.

Postgres 18, RabbitMQ ve RustFS `deploy/third-party.yml` içindedir. İki profil bu dosyayı dahil eder.

## CI

GitHub Actions `push` ve `pull_request` üzerinde üç iş çalıştırır:

- `go test ./...`
- `go run ./tools/fileguard`
- `go run ./tools/archguard`

Birim işi Postgres, RabbitMQ ve RustFS istemez.

## Plan

1. [Kontrol sırası](docs/harpy-starter/00-kontrol-sirasi.md)
2. [Harpy reposuna geçiş](docs/harpy-starter/01-harpy-reposuna-gecis.md)
3. [Starter’ın şekli](docs/harpy-starter/02-starterin-sekli.md)
4. [İlk çalıştırma, tenant profilleri, RustFS](docs/harpy-starter/03-ilk-calistirma-tenant-rustfs.md)
5. [fileguard ve archguard](docs/harpy-starter/04-fileguard-archguard.md)
6. [Uygulama adımları](docs/harpy-starter/05-uygulama-adimlari.md)
