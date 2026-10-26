# Harpy starter — kontrol sırası

Harpy, adını harpi kartalından (harpy eagle) alır. Go ile yazılmış bir backend starter'dır.

Starter'da IAM, iş kuyruğu, worker, single-tenant ve multi-tenant, platform, composition ve `core/cmd/api`, `core/cmd/worker`, `core/cmd/migrator`, `core/cmd/operator`, `cmd/setup` vardır. Yerel üçüncü parti Postgres 18, RabbitMQ ve RustFS'tir. Blob erişimi tek S3 API'dir (`aws-sdk-go-v2`).

Starter'da yoktur: `forms`, `workflow`, AI modülü, MCP sunucusu ve tarihsel SQL migration dökümleri. Migrator kodu starter'dadır. Şema SQL'i Harpy için yazılır.

Harpy kararları:

- Yerel obje servisi RustFS'tir. Uygulama S3 konuşur.
- İlk komut `go run ./cmd/setup`. Setup profil sorar. Docker başlatmaz. Single-tenant ve multi-tenant ayrı dosyalardadır.
- `fileguard` ve `archguard` bu şekli zorlar. `forms`, `workflow`, `ai`, `mcp`, `services/ai` ve `services/mcp` importunu reddeder. Bu paketlerin allowlist'i yoktur.

Okuma sırası:

1. [Harpy reposuna geçiş](01-harpy-reposuna-gecis.md)
2. [Starter’ın şekli](02-starterin-sekli.md)
3. [İlk çalıştırma, tenant profilleri, RustFS](03-ilk-calistirma-tenant-rustfs.md)
4. [fileguard ve archguard](04-fileguard-archguard.md)
5. [Uygulama adımları](05-uygulama-adimlari.md)

Üst `docs/harpy-starter/` ile bu dizin aynı metni taşır. Çelişirse `harpy/docs/harpy-starter/` geçerlidir.
