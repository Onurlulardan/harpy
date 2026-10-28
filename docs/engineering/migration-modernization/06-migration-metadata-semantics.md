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

## Kabul kriterleri

- Her alanın amacı ve runtime etkisi belgelenmiş.
- Sabit değerlerin bilinçli mi eksik mi olduğu kanıtlanmış.
- Geçmiş release uyumluluğu değerlendirilmiş.
- Rollout davranışı için test planı var.
