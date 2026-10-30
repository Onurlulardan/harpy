# Modül — Tenancy

## 1. Owner alanı

Tenancy, tenant kimliği, tenant kayıtları, tenant provisioning/onboarding, deployment policy, installation delivery, tenant settings ve tenant worker lifecycle alanlarının owner'ıdır.

Kaynak kök: core/internal/modules/tenancy/

## 2. Tenant identity

TenantID 16 byte sabit boyutlu değerdir. Parser:
- 36 karakterlik canonical UUID formatı bekler.
- Tire konumlarını doğrular.
- Yalnız lowercase hexadecimal karakter kabul eder.
- Zero UUID'yi reddeder.

Bu katı parse davranışı tenant kimliği string'inin tek bir canonical biçimde kalmasını sağlar.

## 3. Tenant schema naming

SchemaName, tenant ID'den server-side üretilir:
- TenantID base32 encoding kullanılır.
- Prefix t_ eklenir.
- Schema name tenant identity'nin deterministik türevidir.
- Zero tenant ID için schema üretilemez.

SchemaName bir client input değildir. HTTP header, slug veya arbitrary schema string üzerinden SQL search_path kurulmaz. Runtime tenant resolution platform kayıtlarıyla doğrulanır.

Kaynak: domain/tenant_id.go ve domain/schema_name.go.

## 4. Application alt alanları

| Paket | Sorumluluk |
| --- | --- |
| customerprovisioning | SaaS customer/tenant oluşturma operasyonu |
| customerlifecycle | Customer lifecycle command/query ve worker koordinasyonu |
| customerworker | Asenkron customer işlerinin claim/execute modeli |
| onboarding | Tenant onboarding planı, claim, activation ve reconciliation |
| deploymentpolicy | Deployment mode ve tenant provisioning policy |
| installationdelivery | Installation artifact/credential teslimat adımları |
| primarytarget | Primary tenant target çözümleme |
| recoverytarget | Recovery/secondary target çözümleme |
| tenantsettings | Tenant adı ve tenant-level settings güncellemeleri |

## 5. Customer provisioning akışı

CustomerProvisioning Create çağrısı customer request, caller kind, idempotency key ve credential envelope alır.

Provisioning akışında:
1. Caller ve idempotency key validation yapılır.
2. SaaS deployment policy ve control target çözülür.
3. Request digest oluşturulur.
4. Caller capability/access doğrulaması yapılır.
5. SaaS free base offering doğrulanır.
6. Stable provisioning plan hazırlanır.
7. İlk admin hazırlığı için IAM/setup sınırına çağrı yapılır.
8. Operation sonucu persisted state ve receipt ile döndürülür.

Idempotency request digest üzerinden aynı logical request'in tekrarını ayırt eder. Aynı key farklı payload ile kullanılırsa replay olarak kabul edilmemelidir.

## 6. Onboarding ve activation

Onboarding Coordinator plan, claim, accept, activate, reconcile ve recovery credential delivery adımlarını içerir.

Activate:
- Operation ID, TenantID, tenant kind ve plan digest eşleşmesini doğrular.
- Transaction Unit of Work içinde activation claim alır.
- Claim'in operation ve plan ile aynı olduğunu doğrular.
- Tenant schema ve owner binding kaynaklarını açar.
- tenantactivation capability'sine doğrulanmış grant aktarır.
- Callback yalnız activation scope ve doğrulanmış resources ile çalışır.

Activation grant; environment/tenant/first user kimliklerini, offering pin, signing pin, authority pin ve revision/lease bilgilerini taşır. Bu, activation'ın yalnız tek bir boolean switch olmadığını gösterir.

## 7. Deployment policy

Deployment policy runtime, application ve access katmanları SaaS/on-premise davranışının tek bir config değerinden daha kapsamlı doğrulanmasını sağlar. Customer provisioning gibi işlemler, policy mode ve ilgili target aynı anda doğrulanmadan ilerlemez.

## 8. Tenant settings

TenantSettings Update mevcut tenant scope içinde çalışır. Name registry update ve SMTP settings gibi secret içeren mutation'lar ayrı portlara yönlenir. SMTP password request işlenirken memory buffer temizlenir. Update expected revision ve idempotency key/request digest taşır.

## 9. HTTP/worker entegrasyonu

Tenancy HTTP yüzeyi customer provisioning/lifecycle, tenant settings ve deployment policy alanlarını API composition üzerinden sunar. Worker composition customer provisioning ve activation workload'larını diğer owner processor'larıyla birlikte işletir.

## 10. Kaynak haritası

- Tenant primitives: core/internal/modules/tenancy/domain/
- Public API contracts: core/internal/modules/tenancy/api/
- Provisioning: core/internal/modules/tenancy/application/customerprovisioning/
- Onboarding: core/internal/modules/tenancy/application/onboarding/
- Tenant settings: core/internal/modules/tenancy/application/tenantsettings/
- HTTP and worker wiring: core/internal/composition/api/, core/internal/composition/worker/
- Tenant runtime transaction: core/internal/platform/transaction/
- Contracts: contracts/customerprovisioning/, contracts/tenantactivation/, contracts/authority/tenancy_*.json
- Normative skills: skills/data-tenancy/SKILL.md, skills/platform-runtime/SKILL.md
