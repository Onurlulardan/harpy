# 12 — Paylaşılan capabilities ve orchestration

## 1. Capability kavramı

Capability, birden fazla modülün kullanabileceği teknik veya contract-oriented bir yetenektir. Business owner'ın yerine geçmez. Örneğin artifactobject storage işlemlerini soyutlar; belge yaşam döngüsünün owner'ı Documents olarak kalır.

Kaynak kök: core/internal/capabilities/

## 2. Capability kataloğu

| Capability | Sorumluluk |
| --- | --- |
| restquery | Typed query contract, authorization, planning ve execution |
| artifactobject | S3-compatible object read/write/delete/reconciliation API |
| artifactspool | Geçici artifact üretimi için allocation, lease, writer ve janitor |
| credentialdelivery | Credential delivery artifact ve envelope kontratı |
| licensebootstrap | License bootstrap grant/binding capability |
| restoredlicense | Environment restore sonrası license binding capability |
| tenantactivation | Tenant activation grant ve activation scope |
| tabularmapping | Tabular file mapping ve schema/field mapping |
| tabulartransfer | Tabular import/export operation contract ve transfer |

## 3. REST Query

REST Query detaylı HTTP/execution pipeline'ı [HTTP, API ve kontratlar](05-http-api-ve-kontratlar.md) belgesinde açıklanır. Capability modül consumer'lara common typed query yüzeyi verir:
- Field catalog ve reference resolution
- QuerySpec/aggregate validation
- Authorization profile
- Filter/sort/include/aggregate normalization
- Query budget ve planner
- Parameterized execution
- Cursor ve published query execution

Consumer modül, kendi resource/field ownership bilgisini capability'ye sunar. Capability bilinmeyen field/operator'ü sessizce yoksaymaz ve raw SQL yürütmez.

## 4. Artifact Object

artifactobject API, modülün doğrudan AWS SDK kullanmadan object store ile konuşmasını sağlar. S3 adapter grubu:
- Client ve workload identity
- Upload/read/delete
- Object evidence/hash
- Operation identity
- Reconciliation ve ambiguity handling

Store operation sonucu yalnız HTTP success'e dayanmaz. Object metadata, version, digest ve provider receipt üzerinden doğrulanır. Hedef, ambiguous network sonucunda object var/yok kararını kanıtsız tahmin etmemektir.

Documents bu capability'ye kendi file lifecycle/use-case'i içinden bağlanır. Platform veya module domain kodu AWS SDK tiplerini doğrudan kullanmaz.

## 5. Artifact spool

artifactspool kısa ömürlü local artifact üretimi için bounded manager sağlar:
- Workspace/temp allocation
- Lease identity ve expiration
- Writer API
- Janitor cleanup
- Platform OS için allocation implementation'ı

Spool kalıcı business storage değildir. Lease süresi, disk sınırı ve cleanup policy'si explicit olmalıdır. Final artifact gerektiğinde owner module tarafından object store'a taşınır.

## 6. Credential ve license capabilities

credentialdelivery credential artifact, encrypted payload ve delivery reference contract'larını içerir. Bu capability secret'ın public API response veya normal event payload'ında taşınması yerine kontrollü delivery kanalına aktarılmasını sağlar.

licensebootstrap ilk runtime license authority bağlantısında grant/binding operasyonlarını sınırlar. restoredlicense belirli restore/rehydration sonrasında license authority'yi tekrar bağlayan dar kontratı sağlar.

Bu paketler billing/tenancy/platform owner implementasyonlarının yerine geçmez; o owner'ların kullanabileceği typed capability sınırlarını sağlar.

## 7. Tenant activation

tenantactivation API, doğrulanmış activation grant'i ve callback scope'u tanımlar. Activation:
- Tenant/operation identity,
- Environment ve schema binding,
- First user ve root organization,
- Offering/authority pin,
- Revision ve lease
bağlamlarını doğrular.

Tenancy onboarding application, capability'yi doğrulanmış plan/claim ile çağırır. Capability kendi başına tenant provisioning UI veya business domain değildir.

## 8. Tabular mapping/transfer

tabularmapping file column/field mapping tanımları ve validation'ı sağlar. tabulartransfer import/export transfer state ve receipt contract'larını taşır. İlgili consumer'ın data owner'ı, target resource authorization ve transaction kararını vermeye devam eder.

## 9. Orchestration

core/internal/orchestration birden fazla owner/capability arasındaki sıralı operasyonları koordine eder. Önemli paket aileleri:
- systemresources
- systemresourcelayouts
- environmentoperator
- recoveryanchors

Orchestration modüllerin internal persistence'ına erişmek yerine owner API/port ve platform scope kullanır. Operation state, plan hash ve execution result doğrulanabilir olmalıdır.

## 10. Kaynak haritası

- Capabilities: core/internal/capabilities/
- Artifact object: core/internal/capabilities/artifactobject/
- REST Query: core/internal/capabilities/restquery/
- Tabular: core/internal/capabilities/tabularmapping/, tabulartransfer/
- Tenant activation: core/internal/capabilities/tenantactivation/
- Orchestration: core/internal/orchestration/
- Contract packages: contracts/
