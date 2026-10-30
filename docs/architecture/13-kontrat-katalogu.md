# 13 — Contract kataloğu

## 1. Neden ayrı contracts dizini var?

Harpy'de HTTP wire DTO, event payload, runtime config, authority assertion ve storage ownership gibi sözleşmeler business module struct'larından ayrı tutulur. Böylece process'ler ve modüller arası wire format sürümlenebilir, strict validation ve contract testleriyle korunabilir.

Kaynak kök: contracts/

## 2. Ana contract grupları

| Dizin | İçerik |
| --- | --- |
| authority/ | Capability, authority, owner ve operation policy tanımları |
| http/ | HTTP path ve schema/OpenAPI tanımları |
| events/ | Event envelope, event schema, producer/consumer policy |
| runtime/ | Runtime config, health, error, binary ve process contracts |
| storage/ | Storage owner ve DB object ownership contracts |
| asyncmanifest/ | Async producer/consumer/job catalogs |
| jobs/ | Job kind/payload/request contracts |
| restquery/ | Query document and query wire contract |
| actorassertion/ | Signed actor assertion ve peer binding contract |
| credentialdelivery/ | Credential artifact/manifest contract |
| environmentrecovery/ | Recovery artifact contract |
| recoverykit/ | Recovery kit build/parse/validate contract |
| recoverysigning/ | Recovery signing trust bundle contract |
| fcart/ | Encrypted artifact wire format |

## 3. HTTP contracts

HTTP path ve schema dosyaları capability/modül bazında bölünmüştür. Örnek gruplar:
- IAM auth, identity, authorization ve RBAC
- Billing catalog, subscription, offering, invoice, usage, webhook
- Documents
- Work
- Analytics query/dashboard/publication
- Notifications
- Support ve platform operations

Contract testleri schema/path aggregate'i route implementasyonlarıyla eşleştirir. Endpoint değişikliği yalnız Go handler imzasını değiştirmek değildir; request schema, response schema, error contract, authorization ve generated OpenAPI birlikte değerlendirilir.

## 4. Event contracts

events/v1 içinde:
- Event envelope
- Event identity ve sequence
- Producer catalog/policy
- Consumer policy
- Event payload validation
- Dead-letter envelope
- Effect receipt
- Sequence gap
- Bootstrap/operator receipt
tanımları bulunur.

Event type ve version explicit olmalıdır. Consumer'ın hangi producer/version'dan hangi event'i alabileceği policy üzerinden sınırlandırılır. Event payload içinde credential, raw document content veya key material bulunmamalıdır.

## 5. Async manifests

asyncmanifest/v1 paketinde producer ve consumer katalogları ile modül bazlı policy bulunur. Worker startup bu katalogları kullanarak runtime processor registration ve canlı DB uyumluluğunu doğrular. Jobs kontratları ise job kind ve payload shape'lerini tanımlar.

Bu katalogların görevi yalnız runtime registration değildir: unknown producer, unknown event type/version ve yanlış consumer relationship fail-fast olmalıdır.

## 6. Authority contracts

authority/ altında IAM, tenancy, billing, documents, notifications, Work, platform authority/ops ve system resource alanları için declarative policy JSON'ları bulunur.

Bu tanımlar:
- Hangi owner'ın operation'a sahip olduğunu,
- Hangi command/query surface'in açık olduğunu,
- Hangi capability veya authority evidence gerektiğini,
- Hangi actor/tenant scope'ta çalışılabileceğini
makinece doğrulanabilir biçimde tanımlamak için kullanılır.

YAML root manifest yalnız include listesi ve version bilgisini taşır; domain-specific policy tek tek authority JSON dosyalarında tutulur.

## 7. Runtime contracts

runtime/v1 içinde:
- Typed config ve config schema
- Core API/worker/operator config fragments
- Readiness/liveness response schema
- Error envelope
- Release/build version
- Binary schema contract
- Secret/keyring config fragment
- OpenAPI contract fixtures
bulunur.

Runtime config parse edilen environment variable'ların typed settings'e dönüşmesini sağlar. Binary schema contract, uygulama binary'sinin beklediği DB shape ile runtime ortamının uyumunu doğrulayan explicit contract'tir.

## 8. Storage ownership contracts

storage/ fragment ve owner tanımları tablo/schema owner ve grant sınırlarını tarif eder. DB erişim hakkı business module'ün keyfi SQL çalıştırması anlamına gelmez; owner binder ile modül sınırı ve runtime role sınırı birlikte uygulanır.

## 9. Canonical payload ve crypto formatları

canonicaljson/v1 JSON içeriğini deterministik representation'a getirir. fcart/v1 büyük artifact'lar için authenticated encrypted format ve chunk envelope sözleşmesini tanımlar. Password blocklist, key, recovery bundle ve credential artifact formatlarının versioned contract olarak tutulması wire compatibility için önemlidir.

## 10. Contract değişikliği kontrol listesi

- Owner ve consumer açıkça belirli mi?
- Version ve schema shape değişimi backward compatibility ile incelendi mi?
- Unknown field/enum/operator davranışı strict mi?
- Canonical JSON/digest üretimi deterministik mi?
- Contract fixture ve negative test var mı?
- OpenAPI/event/authority aggregate tutarlı mı?
- Generated file provenance kontrolü geçiyor mu?

## 11. Kaynak haritası

- HTTP: contracts/http/
- Events: contracts/events/
- Authority: contracts/authority/
- Runtime: contracts/runtime/
- Storage: contracts/storage/
- Async manifest: contracts/asyncmanifest/
- Contract testler: ilgili contract paketleri ve core/internal/runtime contract testleri
- Normative skills: skills/api-contract/, skills/events-jobs/, skills/modular-architecture/
