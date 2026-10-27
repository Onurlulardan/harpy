# Harpy nedir

Repo, Go backend starter'ıdır. Komut sırası [README](../../README.md) içindedir. Modül sınırı [modular-architecture](../../skills/modular-architecture/SKILL.md) skill'indedir. Dosya boyutu [code-organization](../../skills/code-organization/SKILL.md) skill'indedir.

Go module yolu `github.com/Onurlulardan/harpy`. Go sürümü `1.26`.

## Ne için

Kişi `git clone https://github.com/Onurlulardan/harpy.git` ile klonlar. `go run ./cmd/setup` profili sorar, secret ve env yazar, sonraki komutları basar. Docker başlatmaz.

Secret `deploy/<profil>/.env` ve `deploy/<profil>/secrets/` altına yazılır. `.gitignore` bu yolları dışarıda bırakır. İkinci çalıştırma var olan secret'ı tutar. Yenilemek için:

```text
go run ./cmd/setup --rotate-secrets
```

Soru sormadan varsayılanlarla yazmak için `go run ./cmd/setup --defaults`.

Profil `single-tenant` veya `multi-tenant` olur.

- `single-tenant` yazar: `DEPLOYMENT_MODE=onpremise` (müşteri tenant tavanı 0).
- `multi-tenant` yazar: `DEPLOYMENT_MODE=saas`.

Single-tenant komutları README'dedir. `multi-tenant` aynı sırayı `deploy/multi-tenant/compose.yml` ve `deploy/multi-tenant/.env` ile basar.

Setup şunları sorar:

1. Profil: `single-tenant` veya `multi-tenant`.
2. HTTP portu.
3. Postgres portu, veritabanı adı, kullanıcı. Parola üretilir.
4. Obje depolama: `rustfs` veya `external`. Harici endpoint Amazon S3 ya da MinIO olabilir.

Path-style, RustFS ve MinIO için açıktır. Amazon S3 için kapalıdır.

## Süreçler

| Süreç | İş |
| --- | --- |
| `cmd/setup` | Profili sorar, secret ve env yazar, sonraki komutları basar. |
| `core/cmd/migrator` | `deploy up` ile Harpy SQL migration uygular ve biter. |
| `core/cmd/operator` | `initialize` ile kurulumu açar. |
| `core/cmd/api` | HTTP sunar. |
| `core/cmd/worker` | Kuyruktaki işi çalıştırır. |

Sıra: setup, compose, migrator, operator, api, worker. Migration komutu migrator'dadır.

## Modüller

`core/internal/modules/` altında şunlar vardır:

`analytics`, `audit`, `billing`, `catalog`, `documents`, `iam`, `notifications`, `platformauthority`, `platformops`, `support`, `system_resource_layout`, `tenancy`, `usage`, `work`.

## Kararlar

Yerel obje servisi RustFS'tir. Uygulama tek S3 istemcisi kullanır (`aws-sdk-go-v2`). Amazon S3 ve MinIO aynı istemcinin başka endpoint'idir.

Postgres 18, RabbitMQ ve RustFS `deploy/third-party.yml` içindedir. İmaj digest'i yalnız bu dosyadadır. İki profil dosyası bu yığını dahil eder ve tenancy modunu seçer.

RustFS tek volume kullanır. Komut `/data`. S3 portu `9000`, konsol `9001`. İkisi loopback'tedir. Kimlik `RUSTFS_ACCESS_KEY_FILE` ve `RUSTFS_SECRET_KEY_FILE` ile gelir. Healthcheck `http://127.0.0.1:9000/health` adresine bakar.

`work`, ağaçtaki modüllerden biridir. Pano, görev, yorum ve takvim onun işidir. Ayrıntı: [work](../../skills/work/SKILL.md).

Satırın nasıl ayrılacağı, aktörün tenant'ı nasıl türettiği ve worker'ın işi hangi tenant kapsamında aldığı `tenancy` ve IAM kodundadır.

## Sonraki sayfalar

- [Mimari](01-mimari.md)
- [Yeni modül](02-yeni-modul.md)
- [fileguard ve archguard](03-fileguard-archguard.md)
