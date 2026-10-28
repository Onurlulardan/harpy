# Harpy Migration Sistemi Modernizasyonu — Görev Planı

**Durum:** 01–06 ve 10 için ölçüm ve dar uygulama `baseline.md` içindedir. 07–09 bu akışın dışındadır. Tarihsel release manifestleri elle değişmedi.

## Amaç

Migration sisteminin release geçmişi büyüdükçe artan build/binary maliyetini, startup işini ve bakım yükünü ölçmek; tekrarlanan catalog üretimini tek kaynağa indirmek; mevcut güvenlik ve upgrade garantilerini koruyarak daha küçük, deterministik ve yeniden kullanılabilir bir bundle tasarlamak.

## Başlangıç bulguları

İnceleme main dalındaki commit 5044757b522d95c4e0afc9df3d214d6145b15016 üzerinden yapıldı. Agent uygulamadan önce güncel branch'te bu bulguları yeniden doğrulamalı.

- core/internal/platform/migration/release/ altında 109 .migration-manifest.json dosyası toplam **115,990,137 byte (~116 MB)**.
- Repo içindeki /migrations/ yollarında sayım **2.000 SQL dosyası / 7,711,677 byte (~7.7 MB)** gösterdi. Repo genelindeki SQL sayımı farklı kapsamla yaklaşık 3.270 dosya / 8.3 MB. Agent aynı komutları tekrar çalıştırıp kapsamları ayrı raporlamalı.
- core/internal/platform/migration/load_release_registry.go, Go embed ile tüm arşivi binary'ye dahil ediyor. LoadReleaseRegistry() çağrısında tüm dosyalar okunup doğrulanıyor.
- LoadReleaseRegistry() operator ve migrator composition yollarında çağrılıyor. Maliyet yapısal olarak görülüyor ancak runtime benchmark henüz alınmadı.
- core/internal/composition/api/run.go içindeki loadSchemaContract() fresh migration belgelerini FreshMigrationDocuments() ile topluyor ve schema contract oluşturuyor. API'nin SQL corpus'unu taramasının gerçekten gerekli olup olmadığı doğrulanmalı.
- Fresh manifest üretimi ve SQL seti seçimi birden fazla yerde tekrarlanıyor: core/internal/platform/migration/fresh_registry.go, cmd/setup/release_schema_fresh.go, core/internal/composition/migrator/migration_documents.go, cmd/setup/release_schema_corpus.go.
- İki fresh-manifest üreticisi Phase=expand, TargetDatabaseGeneration=1, NextMinimumRuntimeGeneration=1 değerlerini sabitliyor. Bu metadata'nın gerçek rollout semantiği taşıyıp taşımadığı tasarım sorusu.
- Sistem Goose provider/custom store, platform ve tenant migration'ları, fresh/adopted varyantları, history supersession ve release attestation davranışlarına sahip. Goose'u değiştirmek veya güvenlik kontrollerini kaldırmak bu planın varsayımı değil.

## Görev sırası

1. [01 — Maliyet taban çizgisi ve profiling](01-baseline-and-profiling.md)
2. [02 — Tek canonical bundle üreticisi](02-canonical-bundle-generator.md)
3. [03 — Historical manifestlerin runtime sınırı](03-historical-registry-runtime-boundary.md)
4. [04 — API startup schema contract](04-api-startup-contract.md)
5. [05 — Content deduplication ve varyantlar](05-content-addressing-and-variants.md)
6. [06 — Phase/generation metadata semantiği](06-migration-metadata-semantics.md)
7. [07 — CI ve generated-artifact hataları](07-ci-and-generated-artifacts.md)
8. [08 — Setup idempotency ve secret rotation](08-setup-idempotency-and-secret-rotation.md)
9. [09 — Management listener erişim sınırı](09-management-listener-bind.md)
10. [10 — Dokümantasyon ve regression gate'leri](10-docs-and-regression-gates.md)

## Agent çalışma kuralları

- Önce kodu, testleri ve mevcut docs'u incele; README tek başına kanıt değildir.
- Ölçülmemiş hız/binary iyileşmesini gerçekleşmiş gibi yazma. Önce/sonra ölçümlerinde aynı makine, Go sürümü, build flags ve komutlar kullanılmalı.
- Migration SQL'lerini, release manifestlerini, attestation'ları veya generated hash'leri elle ve seçici biçimde değiştirme; canonical generator ve doğrulama akışını kullan.
- Eski release'lerle upgrade, offline install, rollback, tenant migration, fresh/adopted seçimi ve history supersession davranışları test edilmeden manifest silme.
- Görevleri küçük PR'lara böl. Her PR'da test komutları ve sonuçları raporlansın.
- Bir görev mimari karar gerektiriyorsa seçenekleri, maliyetleri ve kanıtları raporla; kararı sessizce verme.
- Goose'u kaldırmak veya başka migration motoruna geçmek hedef değil.
