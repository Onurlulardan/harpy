# HRY-008 — Operasyonel observability ve alarm kabul kriterlerini tanımla

**Öncelik:** P1  
**Durum:** Planlanacak.

## Amaç
OpenTelemetry entegrasyonunun var olmasını yeterli kabul etmeden, bir Harpy deployment'ında operatörün arızayı teşhis edebilmesini sağlamak.

## Yapılacaklar
- API request rate, error rate ve latency ölçümlerini doğrula.
- PostgreSQL pool kullanımını, connection wait/timeout ve transaction sürelerini görünür kıl.
- Worker job backlog, retry, dead-letter, lease expiry ve event consumer lag ölçümlerini kontrol et.
- Tenant provisioning/migration başarısızlıklarını ve süresini ayrı izlenebilir hale getir.
- Documents object storage upload/download/reconciliation hatalarını ölç.
- Secret, token, PII ve belge içeriği log/trace/metric'e sızmadığını kontrol et.
- Operatör için temel dashboard ve alarm eşiklerini dokümante et.

## Kabul kriterleri
- API, DB, Worker, migration ve object storage için operasyonel sağlık sinyalleri var.
- Bir job backlog veya migration failure senaryosu dashboard/log/trace üzerinden teşhis edilebiliyor.
- Alarm tanımları aksiyon alınabilir ve gürültü üretmeyecek şekilde belgelenmiş.
- Hassas veri redaction kuralları test edilmiş.
