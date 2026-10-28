# Task 01 — Migration maliyetini ölç ve taban çizgisi oluştur

## Neden

116 MB embedded manifest arşivi, SQL kaynaklarından çok daha büyük. Repo boyutu tek başına kullanıcı maliyetini göstermez: binary size, build süresi, startup süresi, heap/allocation ve manifest doğrulama süresi ayrı ölçülmeli.

## İncelenecek alanlar

- core/internal/platform/migration/load_release_registry.go — embedded arşiv ve LoadReleaseRegistry().
- Çağrı yerleri: core/internal/composition/operator/local_operator_pool.go, restored_identity_pool.go, core/internal/composition/migrator/run_schema_executor.go ve run_tenant_migration_executor.go. Güncel çağrı grafiğini doğrula.
- core/internal/composition/api/run.go — API startup schema-contract yolu.
- core/internal/platform/migration/fresh_registry.go ve cmd/setup/release_schema_fresh.go.
- Build/test komutları ve CI workflow'ları.

## Yapılacaklar

1. .migration-manifest.json, migration SQL ve repo toplam dosya boyutlarını ayrı kapsamlarla sayan tekrarlanabilir komut/script belirle.
2. Manifest boyut dağılımını (min/median/p95/max), toplam byte'ı ve tekrar eden içerik/alanların tahmini payını raporla.
3. LoadReleaseRegistry(), fresh catalog/manifest üretimi, API schema-contract startup ve migrator plan/verify başlangıcı için benchmark veya kontrollü ölçüm ekle.
4. Aynı commit/build ortamında binary boyutunu ölç. Embed etkisini izole etmek için yalnız geçici deney build'i kullan; production davranışını bu görevde değiştirmek zorunlu değil.
5. Sonuçları baseline.md dosyasına yaz: komutlar, Go sürümü, OS/arch, commit SHA, medyan/dağılım ve sınırlamalar.
6. Registry'nin aynı proses içinde kaç kez yüklendiğini belirle. Cache düşünülürse yaşam döngüsü ve doğruluk etkisini ayrıca analiz et.

## Kabul kriterleri

- Sayım tekrarlanabilir ve kapsamları açık.
- Süre, memory/allocation ve binary size ayrı raporlanmış.
- Önce/sonra kıyas için sabit komutlar kaydedilmiş.
- Ölçülmeyen fayda iddia edilmemiş.
