# 01 — Harpy reposuna geçiş

## Karar

Starter, `harpy` reposunun `monorepo` dalındadır. Uç `bdfe960004a82d52f873040dce2577a8ca2510a0`. Geçmiş durur. Yeni iş onun üstüne biner. Dal değiştirilmez, tarih silinmez.

Remote `https://github.com/Onurlulardan/harpy.git`.

Üst `docs/harpy-starter/` ile `harpy/docs/harpy-starter/` aynı metni taşır. Çelişirse harpy içindeki metin geçerlidir.

## Görünen iş

Harpy, adını harpi kartalından alan bir Go backend starter'dır. Kişi repoyu klonlar, `go run ./cmd/setup` ile single-tenant veya multi-tenant profilini seçer, basılan komutlarla yığını ve süreçleri çalıştırır.

Uygulama S3 API konuşur (`aws-sdk-go-v2`). Repoyla gelen yerel obje servisi RustFS'tir. Amazon S3 ve MinIO aynı kapının başka endpoint'idir.

`forms`, `workflow`, AI modülü, MCP sunucusu ve tarihsel SQL migration dökümleri starter'da yoktur. IAM, worker, tenancy, platform ve composition starter'dadır. Migrator kodu starter'dadır. Şema SQL'i Harpy için yazılır.

## Kontrol

- [x] Repo `harpy`, dal `monorepo`, geçmiş korunur.
- [x] `forms`, `workflow`, AI modülü, MCP sunucusu ve tarihsel SQL dökümleri yoktur.
- [x] Go module yolu: `github.com/Onurlulardan/harpy`
- [x] Go sürümü: `1.26`
