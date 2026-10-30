# Modül — Analytics

## 1. Owner alanı

Analytics saved query, query version, dashboard, dashboard version, widget, publication approval, public/capability access, export ve query cache alanlarının owner'ıdır.

Kaynak kök: core/internal/modules/analytics/

## 2. Query domain

Query lifecycle active/retired olabilir. Query version state:
- draft
- published
- needs_revalidation
- retired

Published version immutable contract gibi ele alınır. Query version, execution snapshot, dependency listesi ve field/catalog revisions bir arada doğrulanmalıdır.

Query execution client'ın serbest SQL göndermesi değildir. Analytics application REST Query capability üzerinden typed query document, parameter, resource access ve query version bağını kurar.

## 3. Dashboard domain

Dashboard Version:
- DashboardID
- Version number
- Revision
- Published flag
- Visibility
- Widget listesi

NewDraft maksimum 30 widget kabul eder. Widget ID ve position unique olmalıdır. Publish expected revision ister; daha önce yayınlanmış version veya revision mismatch conflict üretir.

Dashboard oluşturma application:
1. Name, visibility, layout document, widget listesi ve idempotency key doğrular.
2. Tenant command scope bind eder.
3. Widget DTO'larını domain modeline dönüştürür.
4. Billing metered resource decision alır.
5. Dashboard ve initial version'ı persistence store'a yazar.
6. Replay ve owner identity'nin decision ile eşleştiğini doğrular.

## 4. Publication release

Publication Release resource version, audience, cache policy, manifest digest, requester, approver, state ve revision taşır.

Temel geçişler:
- Draft → pending approval: Submit
- Pending approval → approved evidence: Approve
- Pending approval/approved → active: Activate
- Active → revoked: Revoke

Approve:
- Expected revision ister.
- Approver'ın requester'dan farklı olmasını şart koşar.
- Step-up auth time'ın güncel olmasını ister; maksimum age 10 dakikadır.

Tenant audience activation'ı ile public/capability audience activation'ı aynı approval koşuluna sahip değildir. Non-tenant audience için approver identity beklenir.

## 5. Query execution

Execute akışı:
1. QueryID ve VersionID UUIDv7 validation.
2. Parameters canonical document decode.
3. Tenant snapshot scope ve execute permission bind.
4. Version snapshot yükleme.
5. Version'ın published olduğunu ve snapshot contract'ının valid olduğunu kontrol.
6. Query dependency setini çözme.
7. REST Query execution contract'ını bind etme.
8. Cache request'i tenant/query/version/binding/query digest/parameter digest ile üretme.
9. Cache policy revision-bound ise cache lookup.
10. REST Query engine ile execute.
11. Result body validation ve eligible cache write.

Cache key sadece query ID'ye dayanmaz. Tenant, query version, dependency/field binding ve parameters digest'i dahil edilir. Authorization/contract generation değiştiğinde eski cache sonucu kullanılamaz.

## 6. Public publication ve capability

Publication audience tenant, public veya capability düzeyinde modellenir. Capability access token raw formda kalıcı saklanmaz; hash ve expiry ile doğrulanır. Public route request'inde tenant identity client'tan seçilmez; publication pointer ve server-side mapping üzerinden resolve edilir.

Public publication için visibility, approval, manifest integrity ve revocation kontrolleri execution öncesinde çalışır.

## 7. Dashboard export

Dashboard export ayrı worker workload'dur. Export:
- Published dashboard/version snapshot'ını alır.
- Query/export dependency ve permission kontrollerini yapar.
- Artifact keyring ve object artifact storage kullanır.
- Job receipt ve status üretir.
- Retention ve artifact cleanup policy'lerini uygular.

Export worker Analytics module owner boundary'sinde kalır; HTTP API export dosyasını request process'inde senkron üretmek zorunda değildir.

## 8. Kaynak haritası

- Domain query: core/internal/modules/analytics/domain/query/
- Domain dashboard: core/internal/modules/analytics/domain/dashboard/
- Domain publication: core/internal/modules/analytics/domain/publication/
- Query application: core/internal/modules/analytics/application/query/
- Dashboard: core/internal/modules/analytics/application/dashboard/
- Publication: core/internal/modules/analytics/application/publication/
- Export: core/internal/modules/analytics/application/dashboardexport/
- Adapters: core/internal/modules/analytics/adapters/
- REST Query capability: core/internal/capabilities/restquery/
- Composition: core/internal/composition/api/ ve worker/
- Contracts: contracts/authority/analytics.json, contracts/http/analytics*, contracts/events/
- Skill: skills/analytics/SKILL.md
