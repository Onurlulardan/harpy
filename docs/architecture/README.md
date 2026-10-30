# Harpy mimari dokümantasyonu

## Bu belge setinin amacı

Bu dokümantasyon, Harpy'nin kaynak kodda mevcut olan mimarisini geliştirici, teknik lider ve operasyon ekibinin takip edebileceği şekilde açıklar. Üst seviye yönlendirme için [Harpy Teknik Dokümantasyon](../harpy-starter/04-teknik-mimari-ve-gelistirici-rehberi.md) sayfası kullanılır.

## Okuma sırası

| Sıra | Belge | Kapsam |
| --- | --- | --- |
| 1 | [Genel bakış](01-genel-bakis.md) | Ürün sınırı, teknoloji seti, repository haritası |
| 2 | [Mimari ve modül sınırları](02-mimari-ve-modul-sinirlari.md) | Modüler monolit, katmanlar, bağımlılık kuralları |
| 3 | [Runtime ve composition](03-runtime-ve-composition.md) | Setup, API, worker, operator ve süreç yaşam döngüsü |
| 4 | [Platform temelleri](04-platform-temelleri.md) | Config, PostgreSQL, transaction, telemetry, health |
| 5 | [HTTP, API ve kontratlar](05-http-api-ve-kontratlar.md) | Route assembly, API contract, hata modeli, REST Query |
| 6 | [Veri ve tenant izolasyonu](06-veri-ve-tenant-izolasyonu.md) | Schema-per-tenant, scope, owner data, Unit of Work |
| 7 | [Event, job ve worker](07-event-job-ve-worker.md) | Outbox, RabbitMQ, DB job, retry, lease, worker |
| 8 | [Güvenlik modeli](08-guvenlik-modeli.md) | Actor, IAM, RBAC, token, tenant güvenliği, secrets |
| 9 | [Deployment ve operasyon](09-deployment-ve-operasyon.md) | Compose, profiller, local stack, health, log ve backup |
| 10 | [Test ve mühendislik standartları](10-test-ve-muhendislik-standartlari.md) | Test türleri, CI, fileguard, archguard |
| 11 | [Geliştirici rehberi](11-gelistirici-rehberi.md) | İlk kurulum, çalışma komutları, katkı akışı |

## Modül belgeleri

| Modül | Kapsam |
| --- | --- |
| [IAM](modules/iam.md) | Identity, authentication, sessions, RBAC, organizations |
| [Tenancy](modules/tenancy.md) | Tenant kimliği, provision, onboarding, settings |
| [Billing](modules/billing.md) | Money, offering, subscription, invoice, entitlement, usage |
| [Catalog](modules/catalog.md) | Signed package, dependency resolution, component lifecycle, menu |
| [Work](modules/work.md) | Board, task, membership, calendar, recurrence |
| [Documents](modules/documents.md) | Folder, item, version, upload, encryption, scan, archive |
| [Notifications](modules/notifications.md) | Inbox, preference, template, provider, delivery |
| [Analytics](modules/analytics.md) | Query, dashboard, publication, ACL, execution |
| [Diğer modüller](modules/diger-moduller.md) | Audit, Usage, Support, System Resource Layout, Platform Authority/Ops |

## Mimariyi doğrulamak için başlıca kaynaklar

- `AGENTS.md`: repository çalışma protokolü ve değişiklik sınırları.
- `skills/INDEX.md`: iş türüne göre zorunlu skill yönlendirmesi.
- `skills/modular-architecture/SKILL.md`: modül sahipliği ve dependency yönü.
- `skills/code-organization/SKILL.md`: dosya organizasyonu ve boyut standardı.
- `skills/testing-quality/SKILL.md`: test seçimi ve kalite kapıları.
- `contracts/`: HTTP, event, runtime ve authority kontratları.
- `core/internal/composition/`: process wiring ve runtime assembly.

## Dokümanların doğruluk sınırı

Modül tanımları gerçek kaynak dizinleri, domain tipleri, application use-case'leri, adapter grupları ve test varlığı üzerinden hazırlanmıştır. Skill dosyalarındaki normatif kurallar ile doğrudan kaynak kodundan gözlenen davranış birbirinden ayrılmıştır. Bir modül içindeki her use-case bu rehberde tek tek listelenmediğinde ilgili modül ağacı esas alınmalıdır.
