# Task 04 — API startup sırasında SQL migration corpus taramasını azalt

## Problem

core/internal/composition/api/run.go içindeki loadSchemaContract() yolu compositionmigrator.FreshMigrationDocuments() çağırıyor ve schemacontract.LoadStartup(documents) ile contract oluşturuyor. API migration çalıştırmıyor; buna rağmen SQL corpus'unu tekrar tarıyor olabilir.

## Yapılacaklar

1. loadSchemaContract() ve schemacontract.LoadStartup altındaki doğrulamaları incele.
2. API'nin ihtiyaç duyduğu alanları listele: manifest digest, schema generation, supported contract version, release identity ve diğer invariant'lar.
3. Bu bilgilerin SQL dosyalarını taramadan build-time üretilmiş küçük schema-contract artifact'ından doğrulanıp doğrulanamayacağını dene.
4. API binary'si ile migrator binary'sinin sorumluluk sınırını belirle. Migration çalıştıran bileşen full bundle taşıyabilir; yalnız contract kontrolü yapan bileşen küçük verified contract taşıyabilir.
5. Contract ile migration bundle arasında cryptographic digest bağı kur. Artifact'ların ayrı ayrı ve uyumsuz güncellenmesini CI engellesin.
6. Contract uyuşmazlığında mevcut fail-closed davranışı koru.
7. Startup testleri ve stale/invalid contract ile fail-closed testi ekle.

## Uygulama

`LoadBoundStartup` gömülü `fresh-migration-manifest.json` baytlarının SHA-256 değerini `binary-schema-contract.json` içindeki `migration_manifest_sha256` ile karşılaştırır. Uyuşmazlık, boş manifest ve doğrulanmamış contract fail-closed olur. SQL taraması `LoadStartup` olarak testte durur. Ölçüm `baseline.md` içindedir.

## Kabul kriterleri

- API startup'ta gereken bilgiler belgelenmiş.
- SQL corpus taramasının azaltılması güvenlik kontrolünü zayıflatmıyor.
- Contract ve bundle arasında doğrulanabilir digest bağı var.
- Startup süresi ve allocations önce/sonra ölçülmüş.
