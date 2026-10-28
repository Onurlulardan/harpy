# Task 06 — Phase ve database/runtime generation metadata'sını netleştir

## Problem

core/internal/platform/migration/fresh_registry.go ve cmd/setup/release_schema_fresh.go entry'leri Phase=expand, TargetDatabaseGeneration=1, NextMinimumRuntimeGeneration=1 olarak sabitliyor. Modelde foundation, expand, backfill, switch, contract fazları bulunuyor. Metadata rollout safety için önemliyse sabit üretim bu anlamı bozabilir; değilse model gereksiz karmaşık olabilir.

## Yapılacaklar

1. Bu alanların bütün okuyucularını, validator'larını ve runtime kararlarını bul. Her alanın etkilediği invariant'ı yaz.
2. Gerçek migration corpus'unda uyumsuz olabilecek entry'leri belirle; yalnız dosya adından phase atama.
3. Seçenekleri raporla: migration başına explicit metadata; SQL yanında sidecar metadata; veya alanları gerçekten gereksizse modelden kaldırma/semantiğini daraltma.
4. Manifest canonicalization, signing, compatibility ve eski manifest okuma etkisini değerlendir.
5. Explicit metadata seçilirse eksik metadata için fail-closed validation ve authoring rehberi ekle.
6. Tüm migration'ları otomatik phase'e çevirmeden önce migration-by-migration karar matrisi üret.

## Karar matrisi

| Alan | Kim okur | Fresh üretici | Arşiv |
| --- | --- | --- | --- |
| `phase` | `Validate` (`foundation`,`expand`,`backfill`,`switch`,`contract`); seçilen migration ve migrator çıktısı taşır | sabit `expand` | 105 foundation, 2615 expand, 15 contract |
| `target_database_generation` | `Validate` gerilemeyi reddeder; `selectReleaseGenerationBoundary` platform maksimumunu alır; history satırı ile karşılaştırılır | sabit `1` | 1..97 arası, `(1,1)` yalnız 105 entry |
| `next_minimum_runtime_generation` | aynı sınır ve history karşılaştırması | sabit `1` | 1..24; binary schema contract generation da 1 |

Sabit `expand/1/1` bilinçlidir: fresh manifest, binary schema contract ile aynı generation 1 sözleşmesidir. Dosya adından phase üretilmez. Arşivdeki gerçek phase ve generation alanları silinmez. Eksik veya bilinmeyen phase `MigrationManifest.Validate` ile fail-closed olur. `cmd/setup/release_postgres_manifests.go` içindeki tek satırlık local fixture `foundation` kullanır; corpus üreticisi değildir.

## Kabul kriterleri

- Her alanın amacı ve runtime etkisi belgelenmiş.
- Sabit değerlerin bilinçli mi eksik mi olduğu kanıtlanmış.
- Geçmiş release uyumluluğu değerlendirilmiş.
- Rollout davranışı için test planı var.
