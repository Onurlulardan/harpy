# Modül — Catalog

## 1. Owner alanı

Catalog, signed package registry, package dependency resolution, component ownership, tenant installation lifecycle ve menu versioning alanlarının owner'ıdır.

Kaynak kök: core/internal/modules/catalog/

## 2. Package manifest contract

Package manifest aşağıdaki bilgileri taşır:
- PackageID, slug ve semantic version
- Platform contract version
- Profile: core veya business
- Required/optional package dependency listesi
- Components
- Menu definition

Manifest canonical byte representation ile saklanır. Validation:
- catalog-package.v1 contract adını ister.
- Slug ve platform contract pattern'lerini kontrol eder.
- Package, component ve menu ID'lerini unique tutar.
- Dependency count, component count ve menu count için üst sınır uygular.
- Component content hash'inin 64 karakter hex olmasını ister.
- Component kind ile owner modülün eşleşmesini doğrular.
- External menu link'lerinde HTTPS ister; user info ve fragment kabul etmez.

Component kind örnekleri FORM, WORKFLOW, DASHBOARD, ENDPOINT, IAM_TEMPLATE ve DEMO_DATA'dır.

## 3. Package registry ve signature

Registry paketi package metadata/manifest çözümleme API'sini sağlar. Package signature doğrulaması key resolver ve canonical payload üzerinden yapılır. Package ID, slug ve version ayrı identity kavramlarıdır.

Dependency resolver required/optional dependency ve semantic version range'lerini birlikte değerlendirir. Geçersiz veya döngüsel dependency grafiği install plan'a çevrilmez.

## 4. Component ownership

Component, component kind'e göre bir owner modüle bağlıdır. Örnek domain eşleşmeleri:
- Workflow component → workflow owner
- Dashboard/endpoint component → analytics owner
- IAM template → IAM owner
- Demo data → catalog/demo owner

Catalog component içeriğini owner modülün domain state'i gibi ele almaz. Lifecycle operation, owner API ile preflight/apply/verify/activate adımlarını koordine eder.

## 5. Installation lifecycle

Lifecycle operation revisioned bir state machine'dir.

Operation kind:
- INSTALL
- UPGRADE
- RETRY
- RESTORE
- RETIRE
- PURGE
- DEMO_INSTALL
- DEMO_REMOVE

Operation state:
- PLANNED
- APPLYING
- WAITING
- VERIFYING
- ACTIVATING
- CANCELLING
- SUCCEEDED
- BLOCKED
- FAILED
- CANCELLED

State değişimi domain transition tablosuyla doğrulanır. Purge için destructive start ayrıca kaydedilir. Operation revision her transition ile artırılır. Cancellation, activation/destructive start veya terminal state sonrasında her zaman kabul edilmez.

## 6. Lifecycle application

Application katmanı:
- Caller/tenant scope ve permission kontrolü yapar.
- Package registry ve signature resolver'ı çağırır.
- Dependency planını kurar.
- Component owner directory üzerinden owner modüllere bağlanır.
- Backup/restore evidence ve operation receipt doğrular.
- Job ve event portlarıyla uzun süren işi worker'a devreder.
- Persisted operation state'i günceller ve idempotent result döndürür.

Uzun süren operation, HTTP request açık tutulmadan worker processor tarafından devam ettirilebilir.

## 7. Menu model

Menu item Group, Form, Workflow, Dashboard, Endpoint ve External Link türlerini tanımlar. Menu version stable ID, version number, revision ve item seti taşır. Item order'a göre deterministic sıralanır; aynı order durumunda stable ID tie-breaker olarak kullanılır.

Tenant override şu alanları değiştirebilir:
- Label
- Order
- Visible

Override item ID'si mevcut published item içinde bulunmalı ve değerler validation'dan geçmelidir. Menu visibility tek başına business permission değildir; owner module authorization her erişimde kontrol edilir.

## 8. Demo ve dependency

Catalog içinde demo handoff/materialization ve dependency alt uygulamaları bulunur. Demo data gerçek owner data boundary'sini atlayarak üretilemez. Dependency read surface package requirement ve installed state'i açıkça sunar.

## 9. Kaynak haritası

- Package contract: core/internal/modules/catalog/domain/packagecontract/
- Operation state machine: core/internal/modules/catalog/domain/lifecycle/
- Menu: core/internal/modules/catalog/domain/menu/
- Registry API: core/internal/modules/catalog/api/packageregistry/
- Component owner API: core/internal/modules/catalog/api/componentowner/
- Lifecycle application: core/internal/modules/catalog/application/lifecycle/
- Registry application: core/internal/modules/catalog/application/registry/
- HTTP/worker integration: core/internal/modules/catalog/adapters/, core/internal/composition/
- Contracts: contracts/catalog*, contracts/events/catalog*, contracts/http/catalog*
- Skill: skills/catalog/SKILL.md
