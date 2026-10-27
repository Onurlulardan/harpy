# 03 — İlk çalıştırma, tenant profilleri, RustFS

## Kişinin akışı

`go run ./cmd/setup` profili sorar, secret ve env yazar, sonraki komutları basar. Docker başlatmaz.

```text
docker compose -f deploy/<profil>/compose.yml up -d
go run ./core/cmd/migrator --env-file deploy/<profil>/.env deploy up
go run ./core/cmd/operator --env-file deploy/<profil>/.env initialize
go run ./core/cmd/api --env-file deploy/<profil>/.env
go run ./core/cmd/worker --env-file deploy/<profil>/.env
```

`single-tenant` `DEPLOYMENT_MODE=onpremise` yazar (müşteri tenant tavanı 0). `multi-tenant` `DEPLOYMENT_MODE=saas` yazar. Süreçler bu env dosyasını `--env-file` ile okur. Yönetim dinleyicisinin varsayılanı `0.0.0.0:9090`dır; RustFS `127.0.0.1:9000` üzerinde kalır.

`deploy up` Harpy fresh şema SQL'i bu ağaçta yokken başarılı olmaz. Eksik sözleşme `binary schema contract is not embedded: expected binary-schema-contract.json` ve `Harpy fresh schema is not present` cümleleriyle stderr'e yazılır. Cutover bu ilk komut değildir.

`--defaults` soru sormaz. Secret git'e girmez. İkinci çalıştırma var olan secret'ı yenile denmeden değiştirmez.

### Sorular

1. Profil: single-tenant veya multi-tenant.
2. HTTP portu.
3. Postgres portu, veritabanı adı, kullanıcı. Parola üretilir.
4. Obje depolama: gömülü RustFS veya harici S3 endpoint. Harici endpoint Amazon S3 ya da MinIO olabilir. Path-style, RustFS ve MinIO için açık, Amazon S3 için kapalıdır.

Cevaplar seçilen profilin env dosyasına yazılır. Süreçler bu env ile açılır.

## Profil dosyaları

Üçüncü parti servisler tek dosyada durur. İki profil bu dosyayı dahil eder. Ayrım tenancy modudur.

```text
deploy/third-party.yml
deploy/single-tenant/compose.yml
deploy/multi-tenant/compose.yml
```

`third-party.yml` Postgres 18, kuyruk için RabbitMQ ve RustFS'i bir kez tanımlar. Digest yalnız bu dosyadadır. MinIO imajı, `minio/mc` ve MinIO env adı yoktur.

Uygulama S3 adları kullanır: endpoint, region, bucket, access key, secret key, path-style. Üç depolama için üç kod yolu yoktur. Blob erişimi tek S3 API'dir (`aws-sdk-go-v2`).

### RustFS

- İmaj digest ile pinlenir. Tek volume, komut `/data`.
- S3 `9000`, konsol `9001`. İkisi loopback'tedir.
- Kimlik `RUSTFS_ACCESS_KEY_FILE` ve `RUSTFS_SECRET_KEY_FILE` ile gelir.
- Access key `A-Z` ve `0-9`. Varsayılan `rustfsadmin` yazılmaz.
- Healthcheck `http://127.0.0.1:9000/health`. API readiness `HeadBucket` yapar.
- Bucket, servis sağlıklı olduktan sonra oluşur.

Upstream RustFS compose'undaki dört disk, Grafana, Prometheus ve Jaeger alınmaz.

## Tenancy

Single-tenant ve multi-tenant, Harpy'nin tenancy kararıdır. Paylaşılan şema, RLS veya `X-Tenant` ile yeni bir model yazılmaz.

Setup yalnız profili seçtirir. Satırın nasıl ayrılacağı, aktörün tenant'ı nasıl türettiği ve worker'ın işi hangi tenant kapsamında aldığı, `tenancy` ve IAM kodundadır.

## Kontrol

- [x] İki profil dosyası tenancy modunu seçer.
- [x] Postgres, RabbitMQ ve RustFS yalnız `deploy/third-party.yml` içindedir.
- [x] İlk komut `go run ./cmd/setup`. `--defaults` soru sormaz.
