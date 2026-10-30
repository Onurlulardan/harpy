# 04 — Platform temelleri

## 1. Platform katmanının yeri

`core/internal/platform` business domain'i değil, process'lerin ve modüllerin ihtiyaç duyduğu teknik altyapıyı sağlar. Bu alanın modül import etmemesi mimari bağımsızlığın önemli bir parçasıdır.

Başlıca paket grupları:

| Alan | Paket |
| --- | --- |
| Config | `platform/config`, `platform/envfile` |
| PostgreSQL | `platform/postgres` |
| Transaction/tenant scope | `platform/transaction` |
| HTTP edge/server | `platform/httpedge`, `httpinput`, `httpserver`, `httpwire` |
| Health | `platform/health`, `platform/management` |
| Telemetry | `platform/telemetry` |
| Async runtime | `platform/async` |
| Process lifecycle | `platform/lifecycle`, `processhost`, `processcode` |
| Identity/actor context | `actorcontext`, `actorbridge`, `requestmeta`, `requestattribution` |
| Schema/runtime contract | `schemacontract`, `binarycontract` |
| Artifact and recovery foundations | `recoveryartifact`, `backupreceipt`, `recoverysigning` |

## 2. Typed configuration

Config alanı environment variable'ları tek tek rastgele package'lerden okumak yerine typed settings nesnelerine dönüştürür. Örnekler:
- Environment: development, production, staging, test.
- Deployment mode: `saas` veya `onpremise`.
- HTTP: read-header timeout, idle timeout, unary timeout, max header bytes ve max JSON body bytes.
- Database: DSN, pool davranışı ve health timing.
- IAM, S3, telemetry, RabbitMQ ve module runtime ayarları ayrı config gruplarında tutulur.

HTTP config'te kaynakta görülen varsayılanlar: header timeout 5 saniye, idle timeout 60 saniye, unary timeout 30 saniye, max header 65.536 byte ve JSON body 1 MiB. Değerlerin aralık kontrolü de loader içinde yapılır.

Config doğrulama kaynakları `core/internal/platform/config/` ve `contracts/runtime/v1/config.schema.json` çevresindedir.

## 3. PostgreSQL runtime

`core/internal/platform/postgres.Open`:
1. İstenen process profilini doğrular.
2. Embedded binary schema contract'in doğrulanmış olmasını ister.
3. pgx pool config'ini oluşturur.
4. Profil bazlı connection/session ayarlarını uygular.
5. Pool hook'larını ve startup preflight collector'ını bağlar.
6. İlk bağlantıyı alarak startup kontrollerini tamamlar.
7. Başlangıç kanıtlarını `Pool` runtime nesnesine koyar.

Pool farklı runtime profillerini tanır: API, worker, operator ve özel teknik executor profilleri. Profil, DB session user/current user ve izinli yetenekler gibi güvenlik özelliklerini belirler. Normal runtime ile privileged owner/bootstrap profilleri aynı credential yüzeyini kullanmaz.

## 4. Transaction coordinator

`core/internal/platform/transaction.Coordinator` modül application katmanlarının tenant veya global transaction sınırını açmasını sağlar.

### Tenant transaction

`InTenant(ctx, request, callback)`:
1. Scope request ve context'i doğrular.
2. Aynı context üzerinde nested scope bulunmadığını kontrol eder.
3. Runtime DB transaction açar.
4. Request identity ve tenant authority doğrulamalarından TenantScope üretir.
5. Callback'i yalnız açık scope ile çalıştırır.
6. Scope'u kapatır ve owner binding'lerini seal eder.
7. Hata/context cancel/panic durumunda rollback uygular.
8. Yalnız başarılı callback ve doğrulanmış commit sonrasında success döndürür.

### Global transaction

`InGlobal`, platform/global data için ayrı bir scope türü sağlar. Tenant scope ile karıştırılmaz. `transaction.Bind[T]`, binder owner bilgisini açık scope üzerinden doğrulamadan owner adapter'ını bağlamaz.

## 5. Unit of Work ilkesi

Application command bir iş değişikliğinin transaction sınırını belirler. Aynı transaction'a dahil edilmesi gereken state, audit, outbox, idempotency receipt ve job intent aynı commit'e katılır.

- HTTP middleware transaction başlatmaz.
- Repository global DB veya tenant slug almaz; scope içindeki DB transaction interface'ini kullanır.
- Commit tamamlanmadan HTTP success üretilmez.
- SMTP, RabbitMQ, AI veya object storage gibi ağ işlemleri açık DB transaction'ında yapılmaz.
- Belirsiz commit sonucu normal başarı gibi yorumlanmaz.

## 6. Health ve readiness

`platform/health` monitor, observation, poller ve snapshot tiplerini içerir. Runtime readiness:
- Database erişimini,
- Session/role preflight'ini,
- Embedded schema contract ile DB'nin uyumunu,
- İlgili process'in owner bağımlılıklarını,
- Admission gate'in açılabilir olmasını
kontrol eder.

Health endpoint'leri dependency durumunu maskelemez. Readiness başarısızsa process iş veya trafik kabulüne hazır sayılmaz.

## 7. Telemetry ve logging

`platform/telemetry.New` build bilgisi, service instance ID ve structured logger ile telemetry runtime kurar. Telemetry kapalıysa OpenTelemetry noop provider kullanılır; açıksa OTLP trace/metric exporter hazırlanır.

İzlenebilirlikte request/correlation ID, actor, tenant, operation, module ve process/service bilgisi kontrollü şekilde taşınır. Secret, token, password veya hassas credential değerleri log/metric/trace attribute olarak yazılmamalıdır. Metric label cardinality'si tenant/user/record ID gibi yüksek çeşitlilikli değerlerle şişirilmez.

## 8. Kaynak kod haritası

- Config: `core/internal/platform/config/`
- PostgreSQL pool/profile: `core/internal/platform/postgres/`
- Tenant/global Unit of Work: `core/internal/platform/transaction/`
- Health: `core/internal/platform/health/`
- Telemetry: `core/internal/platform/telemetry/`
- Process lifecycle: `core/internal/platform/lifecycle/`
- Runtime contract: `contracts/runtime/v1/`
- Runtime wiring: `core/internal/composition/`

Normatif detay: `skills/platform-runtime/SKILL.md` ve `skills/data-tenancy/SKILL.md`.
