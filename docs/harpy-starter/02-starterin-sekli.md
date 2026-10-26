# 02 — Starter’ın şekli

Harpy'nin ağacı bu şekildedir. `forms` ve `workflow` dizinleri yoktur. Yalnız onlara hizmet eden bağ da yoktur. fileguard ve archguard bu şekli zorlar.

## Dizin

```text
harpy/
  cmd/setup/
  core/cmd/api/
  core/cmd/worker/
  core/cmd/migrator/
  core/cmd/operator/
  core/internal/composition/
  core/internal/platform/
  core/internal/modules/iam/
  core/internal/modules/tenancy/
  core/internal/modules/work/
  core/internal/modules/catalog/
  core/internal/modules/documents/
  core/internal/modules/notifications/
  core/internal/modules/analytics/
  core/internal/modules/audit/
  core/internal/modules/billing/
  core/internal/modules/usage/
  core/internal/modules/support/
  core/internal/modules/platformops/
  core/internal/modules/platformauthority/
  core/internal/modules/system_resource_layout/
  deploy/third-party.yml
  deploy/single-tenant/
  deploy/multi-tenant/
  tools/fileguard/
  tools/archguard/
  contracts/
  docs/harpy-starter/
  README.md
```

`core/internal/modules/forms` ve `core/internal/modules/workflow` açılmaz.

## Süreçler

| Süreç | İş |
|---|---|
| `cmd/setup` | Profili sorar, secret ve env yazar, sonraki komutları basar. Docker başlatmaz. |
| `core/cmd/api` | HTTP sunar. Açılışta migration çalıştırmaz. |
| `core/cmd/worker` | Kuyruktaki işi çalıştırır. |
| `core/cmd/migrator` | Harpy için yazılmış SQL migration uygular ve biter. |
| `core/cmd/operator` | Operator süreci. |

## Kodun duruşu

Ayrı bir lifecycle, IAM, tenancy veya kuyruk yazılmaz. `forms`, `workflow`, AI modülü ve MCP sunucusu bağları yoktur.

- HTTP standart kütüphane, log `slog` JSON, sürücü `pgx`, migration Goose.
- Modül yönü: adapter → application → domain. Başka modülün iç paketi import edilmez.
- Platform, modül import etmez. `cmd` süreçleri kendi composition girişini import eder.
- Transaction application tarafında açılır. `TenantScope` durur. Tenancy kararı geçer.
- Blob kapısı tektir ve S3 API konuşur (`aws-sdk-go-v2`). Yerel servis RustFS'tir. MinIO imajı yoktur.

## Starter'da olmayanlar

- `forms` ve `workflow`
- AI modülü ve MCP sunucusu
- Tarihsel SQL migration dökümleri. Migrator kodu starter'dadır. Eski `.sql` dökümü starter şeması değildir. Şema SQL'i Harpy için yazılır.
- Yalnız `forms`, `workflow`, AI modülü veya MCP sunucusuna hizmet eden compiler, DDL, yayın ve runtime bağları

## Kontrol

- [x] Worker, IAM ve tenancy bu şekildedir.
- [x] Örnek modül yoktur. `work`, workflow motoru olmadığı için starter'dadır.
- [x] Go module yolu `github.com/Onurlulardan/harpy`. Go sürümü `1.26`.
