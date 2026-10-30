# 02 — Mimari ve modül sınırları

## 1. Mimari stili

Harpy'nin iş kodu modüler monolit olarak düzenlenmiştir. Modül, business capability için kod ve veri sahipliği sınırıdır; process/deployable sınırı değildir. API ve worker aynı modülün application katmanlarını kullanabilir, fakat her modülün hangi veriye ve kurala sahip olduğu açıktır.

Bu yaklaşım:
- İş mantığını birbirinden bağımsız sahiplik alanlarına böler.
- Her modülün kendi domain modelini ve persistence adapter'ını tutmasını sağlar.
- Modüller arası erişimi dar ve test edilebilir API'lerle sınırlar.
- Bir servis ayırmayı yalnızca güvenlik, kaynak profili, availability veya dış protokol sınırı gerektirdiğinde gündeme alır.

## 2. Modülün tipik dizin yapısı

Gerçek modüller ihtiyaçlarına göre farklılaşır; boş katman açılmaz. Genel şablon:

```text
core/internal/modules/<module>/
  api/                    # Dış consumer kontratı
  domain/                 # Entity, value object, policy, state transition
  application/            # Use-case ve transaction koordinasyonu
  ports/                  # Consumer'a yakın adapter interface'leri
  adapters/
    http/                 # HTTP decoder/handler/presenter
    postgres/             # Persistence implementation
    events/               # Event producer/consumer adapter
  module.go               # Constructor ve public module surface
```

Örneğin Work modülünde `domain/`, `application/boards`, `application/tasks`, `application/calendar`, `ports/`, `adapters/http` ve `adapters/postgres` bulunur. Her modülün dizin ağacı aynı olmak zorunda değildir; gerçek ihtiyaç belirleyicidir.

## 3. Katman sorumlulukları

| Katman | Sorumluluğu | Taşımaması gereken bağımlılık |
| --- | --- | --- |
| Domain | İş invariant'ı, tip, validation ve state transition | HTTP, pgx, AWS SDK, RabbitMQ, framework |
| Application | Use-case, authorization çağrısı, transaction ve side-effect koordinasyonu | Gin/HTTP request tipi, SQL driver, provider SDK |
| Ports | Modülün dış dünyadan beklediği dar sözleşme | Başka modülün private port'u |
| Adapter | HTTP/SQL/event/provider gibi dış protokolün somut implementasyonu | Başka modülün iç domain'i |
| API | Cross-module command/query request/result kontratı | Adapter implementasyonu ve framework tipi |
| Composition | Constructor ve dependency wiring | Business policy ve use-case kararı |
| Platform | Genel teknik primitive | Business modül import'u |

İzin verilen ana bağımlılık yönü:

```text
HTTP adapter ─┐
Postgres adapter ─┼──> application ───> domain
Event adapter ──┘
composition ────> module constructor
module A ───────> module B/api
```

## 4. Cross-module iletişim

Bir modül başka modülün `domain`, `application`, `ports`, `adapters` veya persistence paketini import etmez. Başka modülün tablosuna SQL/JOIN, foreign-key cascade veya ORM relation yazmaz.

Senkron iş gereksiniminde owner modülün public API'si çağrılır. Birden fazla modülün aynı DB transaction'ında değişmesi gerekiyorsa orchestration application coordinator ve ortak Unit of Work sınırı belirler. Commit sonrası asenkron entegrasyon gerekiyorsa versioned outbox event kullanılır.

Aynı process içindeki normal fonksiyon çağrısını HTTP veya RabbitMQ mesajına dönüştürme. Broker, process'ler arası entegrasyon içindir.

## 5. Modül ile process ayrımı

| Kavram | Anlam |
| --- | --- |
| Modül | Business state ve kuralın owner'ı |
| Process | Bir runtime entrypoint'inin yürüttüğü bağımsız çalışma birimi |
| Composition root | Process'e hangi modül ve adapter'ların bağlandığı yer |
| Capability | Birden fazla modülün kullanabileceği ortak teknik yetenek |
| Orchestration | Birden fazla owner/capability'yi iş akışı içinde koordine eden uygulama |

Yeni bir deployable yalnız ekip/klasör büyüklüğü nedeniyle oluşturulmaz. Credential sınırı, farklı ölçekleme profili, hata izolasyonu veya dış protokol zorunluluğu gibi kanıtlanabilir bir gereksinim aranır.

## 6. Platform ve capabilities

`core/internal/platform` teknik primitive'leri içerir: config, clock/ID, PostgreSQL, transaction, telemetry, health, HTTP transport, lifecycle, async runtime, object/blob transport ve process management.

`core/internal/capabilities` daha üst seviyeli, modüller tarafından kullanılan teknik capability'leri barındırır. Önemli örnekler:
- `restquery`: typed query AST, authorization, planning ve execution.
- `artifactobject`: S3 tabanlı artifact/object API.
- `catalogcomponent` ve `catalogprojection`: component owner ve projection sözleşmeleri.
- `credentialdelivery`, `licensebootstrap`, `restoredlicense`: sınırlandırılmış credential/license capability'leri.
- `tenantactivation`: tenant aktivasyonu için doğrulanmış activation grant.

Bir business policy yalnızca teknik olarak ortak diye platforma taşınmaz; ilgili iş modülünde kalır.

## 7. Orchestration

`core/internal/orchestration` çok adımlı platform işlemlerini koordine eder. Örneğin `systemresources` ve `systemresourcelayouts` sistem kaynakları ile presentation akışlarını, `environmentoperator` kurulum/işletim aşamalarını yönetir.

Orchestration kendi başına yeni bir veri owner'ı değildir. İlgili owner modülün kontratını çağırır, adım sırasını ve operation state'ini koordine eder.

## 8. Mimari koruma

- `tools/archguard`, yasak import yönlerini ve cross-module erişimleri denetler.
- `tools/fileguard`, dosya boyutu, adlandırma ve generated provenance kurallarını denetler.
- `skills/modular-architecture/SKILL.md`, modül sınırı için normatif çalışma kaynağıdır.
- `skills/code-organization/SKILL.md`, dosya organizasyonunun normatif kaynağıdır.

## 9. Kaynak kodda takip

Başlangıç noktaları:
- Modüller: `core/internal/modules/<module>/`
- Modül kurulumları: modül kökündeki `module.go` ve ilgili runtime constructor'ları
- API ve worker wiring: `core/internal/composition/api/`, `core/internal/composition/worker/`
- Ortak teknik primitive'ler: `core/internal/platform/`
- Paylaşılan capability'ler: `core/internal/capabilities/`
- Sınır doğrulaması: `tools/archguard/`
