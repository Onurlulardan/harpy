# Starter sınırı

Harpy, adını harpi kartalından (harpy eagle) alır. Go ile yazılmış bir backend starter'dır.

Starter'da vardır: IAM, iş kuyruğu, worker, single-tenant ve multi-tenant, platform, composition, `core/cmd/api`, `core/cmd/worker`, `core/cmd/migrator`, `core/cmd/operator`, `cmd/setup`. Yerel üçüncü parti Postgres 18, RabbitMQ ve RustFS'tir. Blob erişimi tek S3 API'dir (`aws-sdk-go-v2`).

Starter'da yoktur:

- `forms`
- `workflow`
- AI modülü
- MCP sunucusu
- Tarihsel SQL migration dökümleri

Migrator kodu starter'dadır. Şema SQL'i Harpy için yazılır. Eski bir `.sql` dökümü starter şeması değildir.

`work` starter'dadır. Workflow motoru değildir.

Ağaç [starter’ın şekli](harpy-starter/02-starterin-sekli.md) dosyasındadır. İlk komut `go run ./cmd/setup`. Setup profil sorar, Docker başlatmaz, compose / migrate / api / worker komutlarını basar.
