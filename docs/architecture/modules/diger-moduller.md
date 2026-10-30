# Modüller — Audit, Usage, Support ve Platform alanları

Bu belge, ayrı modül dosyası açılmayan kalan owner alanlarını tanımlar. Kod sahipliği her modülün kendi dizininde kalır.

## 1. Audit

Kaynak: core/internal/modules/audit/

Audit, business operation'ların denetlenebilir kayıtlarını tutar. Entry şu bilgileri içerir:
- Audit entry ve owner audit ID
- Actor, organization ve correlation ID
- Owner module ve operation
- Entity type/ID ve action
- Authorization version ve reason code
- Field changes ve event SHA-256
- OccurredAt ve RecordedAt

Action create, update, delete, transition, publish, execute, grant, revoke, restore veya purge olabilir. Field classification public, internal, confidential veya secret değerini alır. FieldChange raw before/after yanında SHA-256 digest taşıyabilir.

Append use-case intent OwnerOperation'ın aktif TenantScope OperationID ile eşleşmesini zorunlu tutar. Entry digest üretilir ve doğrulanmış owner audit store'a yazılır. Böylece Audit dışarıdan keyfi event kaydeden generic log endpoint'i değildir; owner operation bağlamıyla çalışır.

## 2. Usage

Kaynak: core/internal/modules/usage/

Usage HTTP observation toplar ve ürün feature'larına göre sınıflandırır. Record input:
- EnvironmentID, TenantID, ActorID
- Source identity
- HTTP method ve route pattern
- Status code ve duration microseconds
- OccurredAt

Observation canonicalize edilir ve digest'i hesaplanır. OccurredAt gelecekte bir dakikadan fazla veya 24 saatten eski olamaz. Store idempotent observation digest ile kayıt yapar.

FeatureForRoute route pattern'lerini Authentication, Documents, Notifications, Platform Operations, Analytics, Permissions, Administration, Invitations gibi kullanım kategorilerine eşler.

Read application:
- Overview
- Active users
- User summary/timeline
- Trend
- Heatmap
- Top users
- Retention report
alanlarını içerir.

Usage observation'ları business truth değil, kullanım analytics fact'leridir; route pattern ve actor/tenant scope düzgün canonicalize edilmelidir.

## 3. Support

Kaynak: core/internal/modules/support/

Support modülü bug report ve contact request owner'ıdır. BugInput:
- Title, description ve reproduction steps
- Page URL ve user-agent
- Category ve severity
- Browser info
- Console errors
- Network info

Metinler normalize edilir ve alan limitleri uygulanır. JSON browser/console/network payload'ları canonicalizer ile bounded olarak parse edilir; keyfi büyük JSON veya sınırsız error listesi kabul edilmez.

Bug status:
- open
- in_progress
- resolved
- closed
- wont_fix
- duplicate

BugReport tenant/reporter identity, tenant ve reporter snapshot bilgileri, priority, resolution/admin notes, revision ve timestamps taşır. Tenant user report ile platform support operations ayrı access path'leri kullanır.

## 4. System Resource Layout

Kaynak: core/internal/modules/system_resource_layout/

Bu modül sistem kaynaklarının layout/form-like definition ve presentation catalog'unu yönetir. Domain-specific forms engine olarak genellenmemelidir; mevcut kaynakta system resource presentation owner'ıdır.

Application yüzeyi:
- Definition create/get/list/update/delete
- Publish/presentation revision
- Installation default ve catalog
- Reference resolution
- Runtime hydration ve presentation compile

Publish akışı expected row revision ister, source definition SHA-256 değerini alır, field catalog contract ile presentation document'ını compile eder ve canonical JSON/SHA-256 üretir. Published record compiler version, field catalog contract, presentation contract ve digest taşır. Persist edilen sonuç tekrar doğrulanır.

Presentation field identity, label, section, surface ve widget metadata'sı typed field catalog ile çözülür; client'ın verdiği arbitrary field key kabul edilmez.

## 5. Platform Authority

Kaynak: core/internal/modules/platformauthority/

Bu modül environment/platform authority için signed proposal, acknowledgement, binding proof, checkpoint ve trust bundle kavramlarını yönetir.

Domain artifact alt paketi:
- Proposal ve signed proposal
- Ack ve verified ack
- Binding proof
- Checkpoint
- Trust bundle resolution

Application Coordinator store, registry, signer, clock ve verified trust bundle ister. Flow değişiklik önerisini hazırlar, sign eder, delivery/acknowledgement receipt'lerini doğrular ve signer lifecycle'ını yönetir. Air-gap ve online expiry policy ayrımı vardır.

Adapter grupları PostgreSQL, GCP KMS, local key, air-gap ve provider registry yüzeyleridir.

## 6. Platform Ops

Kaynak: core/internal/modules/platformops/

Platform Ops, operational diagnostics ve recovery operasyonlarının owner'ıdır:
- Alert rules/events
- Trace list/detail/timeline/statistics
- Permission cache flush
- System health
- Recovery set, manifest ve recovery kit
- Restore plan ve restore execution

Recovery manifest/plan canonical JSON olarak üretilir; strict decoder unknown field'ları reddeder. Restore request:
1. Protected recovery set'i bulur.
2. Manifest signature/digest doğrular.
3. Plan signature, operation ID ve artifact binding doğrular.
4. Target environment identity'yi çözer.
5. Safety evidence ve target disposition gate'ini doğrular.
6. Recovery operation oluşturur ve receipt kaydeder.

Backup crypto, recovery storage ve signer adapter'ları domain/application katmanından ayrıdır.

## 7. Modül kaynak haritası

| Modül | Domain | Application | Adapter |
| --- | --- | --- | --- |
| Audit | modules/audit/domain | modules/audit/application | modules/audit/adapters |
| Usage | modules/usage/domain | modules/usage/application | modules/usage/adapters |
| Support | modules/support/domain | modules/support/application | modules/support/adapters |
| System Resource Layout | modules/system_resource_layout/api | modules/system_resource_layout/application | modules/system_resource_layout/adapters |
| Platform Authority | modules/platformauthority/domain | modules/platformauthority/application | modules/platformauthority/adapters |
| Platform Ops | modules/platformops/domain | modules/platformops/application | modules/platformops/adapters |

## 8. Skill ve contract kaynakları

- Audit ve audit policy: contracts/authority/ ve contracts/events/
- Usage: core/internal/modules/usage/ ve contracts/authority/
- Support: core/internal/modules/support/ ve contracts/http/
- System resource layout: contracts/authority/system_resource_layout.json
- Platform authority: contracts/authority/platformauthority.json
- Platform ops: contracts/authority/platformops.json
- Recovery contracts: contracts/environmentrecovery/, contracts/recoverykit/, contracts/recoverysigning/
