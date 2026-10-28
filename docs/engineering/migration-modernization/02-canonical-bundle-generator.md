# Task 02 — Migration bundle üretimini tek canonical kaynağa indir

## Problem

Migration setlerinin hangi path/package'lerden toplandığı ve fresh manifestin nasıl seçilip canonical JSON/hash'e dönüştürüldüğü birden fazla yerde tanımlı. Bu, üreticilerin farklı migration seçmesine veya farklı digest üretmesine neden olabilir.

## İlgili dosyalar

- core/internal/platform/migration/fresh_registry.go
- cmd/setup/release_schema_fresh.go
- core/internal/composition/migrator/migration_documents.go
- cmd/setup/release_schema_corpus.go
- contracts/postgres/v1/migration_types.go ve manifest validation kodu.

## Yapılacaklar

1. İki üreticinin çıktısını karşılaştır: entry seti, sıralama, path, owner/scope, version, transaction mode, phase, generation, variant path/digest ve canonical JSON byte'ları.
2. SQL seti listelerinin farklarını tabloyla çıkar: root, glob, owner/scope ve supersession kuralları.
3. Go internal import sınırlarını dikkate al. cmd/setup repo kökünde olduğundan core/internal paketini doğrudan import edemez. Repo kökündeki internal/migrationbundle olası bir konumdur; nihai yeri kanıtla seç.
4. Tek deterministic API tasarla: SQL kaynaklarını toplama, seçim/supersession, manifest oluşturma, canonical serialization ve digest. Build-time generation ile runtime execution sorumluluklarını ayır.
5. Mevcut iki yolu geçici olarak aynı fixture üzerinde çalıştırıp byte-for-byte eşitliği doğrula; sonra çağıranları ortak implementation'a geçir.
6. Generation komutu ve CI'da “generated output is up-to-date” kontrolü ekle.
7. İki corpus bilinçli olarak farklıysa zorla birleştirme; farkların amacını isim ve testle koru.

## Kabul kriterleri

- Aynı input seti tek ve deterministic manifest/digest üretir.
- Kopya implementasyon kalmaz; bilinçli farklılıklar belgelenir.
- Path ve supersession kararları test edilir.
- Generated manifest drift'i CI'da anlaşılır hata verir.
- Release attestation formatı veya imza doğrulaması sessizce değişmez.
