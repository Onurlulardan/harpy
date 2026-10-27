# 05 — Uygulama adımları

Bu sıra `harpy` `monorepo` dalında geçerlidir. Geçmiş silinmez.

Starter'da yoktur: `forms`, `workflow`, AI modülü, MCP sunucusu ve tarihsel SQL migration dökümleri. Migrator kodu starter'dadır. Şema SQL'i Harpy için yazılır.

## Adım 1 — Plan

- Bu dizin `harpy/docs/harpy-starter/` altındadır. Üst `docs/harpy-starter/` aynı metni taşır.
- `harpy/README.md` Harpy'yi harpi kartalından gelen Go backend starter olarak tanımlar. İlk komutu 03'e bağlar.

Bitti: plan Harpy içindedir.

## Adım 2 — Sınır

Sınır [starter sınırı](../kesim-haritasi.md) dosyasındadır.

- `forms`, `workflow`, AI modülü ve MCP sunucusu starter'da yoktur.
- Tarihsel SQL migration dökümleri starter şeması değildir.
- IAM, tenancy, worker ve `work` starter'dadır.

Bitti: sınır metni starter'da kalanları ve dışarıda kalanları yazar.

## Adım 3 — Süreçler ve modüller

Sıra: platform ve `cmd` süreçleri, tenancy, IAM, kuyruk ve worker, sonra `forms` ve `workflow`'a bağlı olmayan diğer modüller. Yerine örnek modül yazılmaz.

Modül dizinleri `forms`, `workflow`, `ai` ve `mcp` yoktur. `go test ./...` ağacın tamamında kırmızıdır.

## Adım 4 — Profiller ve RustFS

- `deploy/third-party.yml` Postgres 18, RabbitMQ ve RustFS'i bir kez tanımlar.
- İki profil dosyası bu yığını dahil eder ve tenancy modunu seçer.
- `cmd/setup` 03'teki soruları sorar. MinIO imajı yoktur.

Bitti: profil dosyaları ve `cmd/setup` yazıldı. Yığın bu turda ayağa kalkmadı.

## Adım 5 — fileguard ve archguard

04'teki sözleşme Harpy ağacını zorlar. forms, workflow, AI ve MCP allowlist'i yoktur. Araç paketlerinin testleri yeşildir. `go run ./tools/fileguard` ve `go run ./tools/archguard` bu ağaçta çıkar kodu 0 ile biter.

## Adım 6 — README ve CI

README klon, setup, compose, migrate, operator initialize, api ve worker sırasını yazar. CI `go test ./...`, fileguard ve archguard çalıştırır. Üçü de Harpy modülünde çalışır.

Bitti: README klon ve komut sırasını yazar. CI üç işi tanımlar.
