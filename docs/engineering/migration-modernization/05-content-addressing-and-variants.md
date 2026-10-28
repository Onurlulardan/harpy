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

## Kabul kriterleri

- Fresh/adopted corpus kapsamı doğrulanmış.
- Dedup potansiyeli gerçek hash ölçümüyle raporlanmış.
- Origin ve authorization semantiği korunmuş.
- Paylaşılan içeriklerin runtime eşdeğerliği test edilmiş.
- Kazanç yetersizse değişiklik yapmamak geçerli sonuç.
