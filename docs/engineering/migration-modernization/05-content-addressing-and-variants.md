# Task 05 — Fresh/adopted varyantlarını ve content deduplication olanağını incele

## Problem

Release manifest entry modeli variants.fresh ve variants.v1_adopted içeriyor. Her varyant path ve SHA256 taşıyor. Mantıksal origin farkı korunurken aynı byte içeriğinin birden fazla yerde tutulup tutulmadığı ölçülmeli. Path/origin validation kuralları farklı olabilir.

## Yapılacaklar

1. Fresh/adopted SQL dosyalarının gerçek dağılımını doğrula. Önceki incelemede /migrations/ altında v1_adopted path'i görünmemişti; başka roots veya generated release corpus nedeniyle bu sonuç eksik olabilir.
2. Her manifestte aynı SHA256'ya sahip varyantları ve aynı içeriğin farklı path'lerde tekrarını bul.
3. contracts/postgres/v1/migration_types.go ve validation kodundaki path suffix/origin kurallarını incele. Mantıksal varyant kimliği, fiziksel blob kimliği ve manifest referansını ayırmanın etkisini değerlendir.
4. Content-addressed blob store/build artifact dedup seçeneğini gerçek tekrar oranı ve ölçülen kazançla değerlendir.
5. Fresh/adopted execution davranışlarının eşit olduğunu varsayma. Schema state, history row, transaction mode, ordering ve upgrade testlerini belirle.
6. Gerekirse geçişte önce eski path/digest referanslarını doğrula; eski release manifest formatını geriye dönük değiştirme.

## Ölçüm

`go run ./internal/migrationbundle/cmd/census .` (ayrıntı `baseline.md`):

- Repo içinde `*.v1_adopted.sql` yok. Adopted path'ler yalnız arşiv manifestlerinin `variants.v1_adopted` alanındadır.
- 2735 benzersiz entry'nin tamamında fresh ve v1_adopted vardır. 2714 aynı SHA256 ve farklı path. 21 farklı SHA256. Aynı path: 0.
- 21 farktan 19'u `ai` modülü path'i, biri `tenancy` platform `20260713030300`, biri `platform_cutover` `20260716020103`. Bu ağaçta adopted SQL yok; forms, workflow ve AI kaynakları starter'da yoktur.
- Benzersiz entry JSON yaklaşık 1.7 MB. Kazanç, 116 MB arşivin entry tekrarıdır. Bunu content-addressed blob'a çevirmek release attestation baytlarını değiştirir.

Kod değişikliği yok. Fresh ve adopted yürütmesi eşit kabul edilmedi. Goose, history satırı ve origin seçimi duruyor.

## Kabul kriterleri

- Fresh/adopted corpus kapsamı doğrulanmış.
- Dedup potansiyeli gerçek hash ölçümüyle raporlanmış.
- Origin ve authorization semantiği korunmuş.
- Paylaşılan içeriklerin runtime eşdeğerliği test edilmiş.
- Kazanç yetersizse değişiklik yapmamak geçerli sonuç.
