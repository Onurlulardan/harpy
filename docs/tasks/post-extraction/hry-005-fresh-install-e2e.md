# HRY-005 — Single-tenant ve multi-tenant fresh-install E2E senaryolarını doğrula

**Öncelik:** P1  
**Durum:** Planlanacak.

## Amaç
Harpy'nin boş ve disposable bir ortamda, eski local volume veya mevcut DB durumuna güvenmeden kurulabildiğini kanıtlamak.

## Yapılacaklar
- PostgreSQL 18, RabbitMQ ve S3-compatible object storage içeren temiz bir test ortamı hazırla.
- Single-tenant kurulumda setup, baseline, operator initialize, IAM bootstrap, API/Worker readiness ve graceful shutdown akışını doğrula.
- Multi-tenant kurulumda aynı akışa tenant provisioning, ikinci tenant oluşturma ve tenant izolasyon kontrollerini ekle.
- Temel dikey akış olarak IAM authentication/authorization, Work task CRUD ve Documents upload/download senaryolarını çalıştır.
- Worker tarafında en az bir kalıcı job ve bir event delivery akışını doğrula.
- Aynı senaryoları desteklenen deployment profilleri için ayrı ayrı raporla.

## Kabul kriterleri
- Single-tenant ve multi-tenant temiz kurulumları ayrı ayrı başarılı.
- API ve Worker readiness kontrolleri geçiyor.
- İkinci tenant oluşturulması ilk tenant verisini veya erişimini etkilemiyor.
- Document upload/download ve temel authorization akışları çalışıyor.
- Process shutdown sırasında aktif işlerin ve DB bağlantılarının kapanışı kontrollü.
- Test hiçbir gerçek DB veya kullanıcı volume'unu resetlemiyor.
