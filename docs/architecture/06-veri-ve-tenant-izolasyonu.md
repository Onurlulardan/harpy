# 06 — Veri modeli ve tenant izolasyonu

## 1. Data ownership

PostgreSQL, Harpy'nin canonical transactional data store'udur. Her business module kendi tablo ve veri modelinin sahibidir. Başka modülün tablosuna doğrudan SQL/JOIN veya ORM ilişki yoluyla erişilmez.

Cross-module read/write ihtiyaçları:
- Owner modülün public query/command API'si,
- Batch enrichment,
- Versioned integration event,
- Rebuild edilebilir projection
üzerinden karşılanır.

Redis/cache veya RabbitMQ mesajları canonical business state yerine geçmez. Bir projection, sahibi olduğu kaynağa bağlı ve yeniden üretilebilir olmalıdır.

## 2. Platform schema ve tenant schema

Harpy tenant verisini PostgreSQL schema-per-tenant modeliyle ayırır. Tenant kimliği ile fiziksel schema adı ayrı kavramlardır.

- `TenantID`, 16 byte canonical UUID değeridir.
- Parser canonical lowercase UUID biçimi bekler ve sıfır UUID'yi reddeder.
- Tenant schema adı `SchemaNameForTenant` ile server-side üretilir.
- Schema adı istemciden gelen slug/header değeriyle request başına kurulmaz.
- Kaydedilmiş immutable `schema_name` platform registry'den doğrulanır.
- Tek bir transaction bir tenant schema'sı içinde çalışır.
- Platform tablolarına erişim fully-qualified isimle yapılır.

Kod: `core/internal/modules/tenancy/domain/tenant_id.go`, `schema_name.go`.

## 3. TenantScope

TenantScope, sıradan bir tenant ID parametresi değildir. Açık DB transaction'ına, doğrulanmış actor/tenant authority'sine, operation identity'ye ve scope moduna bağlı bir capability'dir.

Application command transaction coordinator üzerinden scope alır. Repository/owner binding yalnız açık ve owner ile eşleşen scope'ta kullanılabilir. Scope callback sonrasında kapatılır; dışarı taşınan stale scope kullanılamaz.

Bu tasarım, tenant kimliğinin caller tarafından seçilmesi yerine verified actor context ve server-side kayıt üzerinden belirlenmesini sağlar.

## 4. Unit of Work sınırı

Tipik mutation akışı:

```text
HTTP / worker command
  -> actor + permission + tenant validation
  -> TenantScope transaction
      -> owner module application
      -> repository reads/writes
      -> audit record
      -> idempotency receipt
      -> outbox event / DB job intent
  -> commit
  -> HTTP response veya async acknowledgement
```

Transaction açılması HTTP middleware'e bırakılmaz. Application command işin atomic boundary'sini belirler. Commit öncesi başarı cevabı verilmez.

## 5. Transaction ve external side effect

Açık DB transaction içinde uzun süreli network I/O yapılmaz. Özellikle:
- S3 upload/download,
- SMTP veya Telegram gönderimi,
- RabbitMQ publish,
- Harici provider çağrısı
DB transaction'ının içinde yürütülmez.

Bu işlemlerin öncesi/sonrası persisted state ve durable receipt ile koordine edilir. Örneğin belge yükleme, object storage işlemi ile tenant DB kaydını tek bir açık transaction gibi ele almaz; upload state ve cleanup receipt ile belirsiz sonuçları yönetir.

## 6. DB connection/session hijyeni

`platform/postgres` runtime pool:
- Process profiline göre role/session preflight yapar.
- Tenant ve global transaction scope'larında search path'i kontrollü kurar.
- Transaction bağlarını seal eder.
- Scope dışına çıkan veya tainted connection'ı normal connection gibi tekrar kullanmaz.
- Readiness sırasında database role ve schema contract uyumunu kontrol eder.

Tenant transaction `search_path` değerinde yalnız `pg_catalog` ve ilgili tenant schema'sı yer alır. Client-controlled schema/table identifier kabul edilmez.

## 7. Platform vs tenant verisi

Platform düzeyindeki veriler installation/environment, deployment mode, tenant registry, global authority ve platform operation gibi tenant seçimi öncesi veya tüm tenantları kapsayan kayıtlardır.

Tenant düzeyindeki veriler bir tenant'a ait kullanıcı, board, belge, notification, billing subscription ve analytics gibi iş verileridir. Her modülün hangi storage scope'u kullandığı ilgili adapter ve contract üzerinden belirlenir.

## 8. Data model sınırlarının örnekleri

| Alan | Owner | Model örnekleri |
| --- | --- | --- |
| Identity | IAM | User, Role, Organization, Profile |
| Tenant registry/settings | Tenancy | TenantID, SchemaName, deployment policy |
| Money and subscriptions | Billing | Money, Period, Subscription, Entitlement |
| Work management | Work | Board, Column, Task, Membership |
| Files | Documents | Folder, Item, Version, Upload |
| Inbox/delivery | Notifications | Message, Preference, Delivery |
| Saved analytics | Analytics | Query Version, Dashboard Version, Publication |
| Audit | Audit | Entry, FieldChange, Classification |

## 9. Kaynak kod haritası

- Tenant ID/schema: `core/internal/modules/tenancy/domain/`
- Transaction scopes: `core/internal/platform/transaction/`
- PostgreSQL pool/owner binding: `core/internal/platform/postgres/`
- Module-owned adapters: `core/internal/modules/*/adapters/postgres/`
- Storage ownership contracts: `contracts/storage/`
- Tenant runtime contracts: `contracts/runtime/v1/`

Normatif prosedür: `skills/data-tenancy/SKILL.md` ve `skills/modular-architecture/SKILL.md`.
