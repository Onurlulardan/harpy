# Harpy Mimari Dokümantasyonu

## Amaç

Bu dokümantasyon, Harpy'nin kaynak kodda implement edilmiş mimarisini geliştirici, teknik lider ve operasyon ekiplerinin takip edebileceği şekilde açıklar. Teknik mimari için ana başvuru kaynağı bu dizindir; proje özeti ve hızlı başlangıç yönergeleri [kök README](../../README.md) dosyasındadır.

## Okuma sırası

| Sıra | Belge | Kapsam |
| --- | --- | --- |
| 1 | [Genel bakış](01-genel-bakis.md) | Ürün kapsamı, teknoloji seti, repository haritası |
| 2 | [Mimari ve modül sınırları](02-mimari-ve-modul-sinirlari.md) | Modular monolith, katmanlar, bağımlılık kuralları |
| 3 | [Runtime ve composition](03-runtime-ve-composition.md) | Setup, API, worker, operator ve süreç yaşam döngüsü |
| 4 | [Platform temelleri](04-platform-temelleri.md) | Config, PostgreSQL, transaction, telemetry, health |
| 5 | [HTTP, API ve contracts](05-http-api-ve-kontratlar.md) | Route assembly, API contract, hata modeli, REST Query |
| 6 | [Veri ve tenant izolasyonu](06-veri-ve-tenant-izolasyonu.md) | Schema-per-tenant, scope, data ownership, Unit of Work |
| 7 | [Event, job ve worker](07-event-job-ve-worker.md) | Outbox, RabbitMQ, DB job, retry, lease, worker |
| 8 | [Güvenlik modeli](08-guvenlik-modeli.md) | Actor, IAM, RBAC, token, tenant güvenliği, secrets |
| 9 | [Deployment ve operasyon](09-deployment-ve-operasyon.md) | Compose, profiller, local stack, health, log ve backup |
| 10 | [Test ve mühendislik standartları](10-test-ve-muhendislik-standartlari.md) | Test türleri, CI, fileguard, archguard |
| 11 | [Geliştirici rehberi](11-gelistirici-rehberi.md) | İlk kurulum, çalışma komutları, katkı akışı |
| 12 | [Paylaşılan capabilities](12-paylasilan-capabilities.md) | REST Query, artifact, tenant activation ve orchestration |
| 13 | [Contract kataloğu](13-kontrat-katalogu.md) | HTTP, event, runtime, authority ve storage contracts |
| 14 | [Mimari kararlar](14-mimari-kararlar.md) | Modular monolith ve önemli trade-off'lar |
| 15 | [Modül paket haritası](15-modul-paket-haritasi.md) | 13 modülün application ve adapter alt paketleri |

## Modül belgeleri

| Modül | Kapsam |
| --- | --- |
| [IAM](modules/iam.md) | Identity, authentication, sessions, RBAC, organizations |
| [Tenancy](modules/tenancy.md) | Tenant kimliği, provisioning, onboarding, settings |
| [Billing](modules/billing.md) | Money, offering, subscription, invoice, entitlement, usage |
| [Work](modules/work.md) | Board, task, membership, calendar, recurrence |
| [Documents](modules/documents.md) | Folder, item, version, upload, encryption, scan, archive |
| [Notifications](modules/notifications.md) | Inbox, preference, template, provider, delivery |
| [Analytics](modules/analytics.md) | Query, dashboard, publication, ACL, execution |
| [Diğer modüller](modules/diger-moduller.md) | Audit, Usage, Support, System Resource Layout, Platform Authority/Ops |

## Mimariyi doğrulamak için başlıca kaynaklar

- `AGENTS.md`: Repository çalışma protokolü ve değişiklik sınırları.
- `skills/INDEX.md`: İş türüne göre zorunlu skill yönlendirmesi.
- `skills/modular-architecture/SKILL.md`: Modül sahipliği ve dependency yönü.
- `skills/code-organization/SKILL.md`: Dosya organizasyonu ve boyut standardı.
- `skills/testing-quality/SKILL.md`: Test seçimi ve kalite kapıları.
- `contracts/`: HTTP, event, runtime ve authority contracts.
- `core/internal/composition/`: Process wiring ve runtime assembly.

## Dokümanların doğruluk sınırı

Modül tanımları gerçek kaynak dizinleri, domain tipleri, application use-case'leri, adapter grupları ve test varlığı üzerinden hazırlanmıştır. Skill dosyalarındaki normatif kurallar ile doğrudan kaynak kodundan gözlenen davranış birbirinden ayrılmıştır. Bir modül içindeki tüm use-case'ler burada tek tek listelenmiyorsa ilgili modülün source tree'si esas alınmalıdır.
