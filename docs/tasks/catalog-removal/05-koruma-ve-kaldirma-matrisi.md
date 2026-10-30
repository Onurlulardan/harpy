# 05 — Koruma ve kaldırma matrisi

## 1. Kaldırılacak ForgeCRUD Catalog yüzeyi

| Alan | Kaldırılacak içerik |
| --- | --- |
| Business module | core/internal/modules/catalog/** (domain, application, API, HTTP, PostgreSQL, worker, demo, adoption, backup, migration ve tests) |
| Runtime | API Catalog HTTP builder/access/readiness/routes; Worker Catalog processor/owner directory |
| IAM integration | Catalog endpoint permissions, tenant release capability values, IAM_TEMPLATE package owner ve Catalog projection adapter/migration'ları |
| Analytics integration | Catalog component/legacy owner, catalogowner/cataloglifecycle adapters, Catalog payload ve Catalog-specific migration/tests |
| Shared capabilities | catalogcomponent ve catalogprojection; ancak son tüketici temizlendikten sonra |
| Package contracts | contracts/catalogpackage/ ve contracts/catalogcomponent/ |
| HTTP contracts | Catalog-v2 paths/schemas ve Catalog root API path fragment |
| Authority/operations | Catalog-specific authority JSON ve operations fragments |
| Events/jobs | catalog.installation.* ve catalog.menu.* schema/policy, Catalog-specific job processor registration |
| Release/setup | Local ForgeCRUD package/demo payload, package set signature/verification, Catalog keyring ve deploy evidence |
| Storage/migrations | Catalog storage owner, migration stream, kendi platform/tenant migration'ları, Catalog owner SQL/routines ve onaylanmış cleanup objects |
| Parity inventory | contracts/catalog/ içindeki ForgeCRUD first-party package/parity manifests |
| Documentation/skill | Catalog architecture module guide, README/architecture module tanımları, final aşamada skills/catalog ve INDEX route |

## 2. Mutlaka korunacak generic catalog kavramları

| Sahip | Korunacak path / kavram | Neden |
| --- | --- | --- |
| PostgreSQL | pg_catalog, search_path ve DB object catalog query'leri | PostgreSQL sistem metadata'sı |
| REST Query | core/internal/capabilities/restquery/** içindeki field/relation/query catalog | Dynamic query field validation ve authorization |
| Field catalog | core/internal/platform/fieldcatalog/** | Typed field identity, schema ve presentation altyapısı |
| Billing | core/internal/modules/billing içindeki offering, catalogquery, discountcatalog; contracts/http/billing-catalog-*; billing_catalog_* authorities/operations | Billing commercial catalog ve ticari fiyat/ürün modeli |
| IAM | Resource/action/permission catalog, permission request catalog, RBAC, domain/platformiam/capability_catalog.go ve system resource catalog | IAM authorization ve platform capability sahipliği |
| System Resource Layout | core/internal/modules/system_resource_layout/application/catalog.go; orchestration/systemresources/catalog.go | Harpy system resource/presentation listing; ForgeCRUD package registry değildir |
| Work | core/internal/modules/work application/restquery ve adapters/postgres/storagecatalog | Work REST Query field/storage mapping |
| Async platform | core/internal/platform/async event_catalogs.go, release_catalogs.go, write_catalogs.go, jobexecution/policy_catalog.go; contracts/events/v1 generic ProducerCatalog/ConsumerCatalog ve contracts/jobs/v1 Catalog | Generic producer/consumer/job registry abstractions |
| Migration platform | core/internal/platform/migration catalog files, tenant_catalog.go; tools/internal/migrationstreamcatalog | Generic migration source/stream registration |
| Storage governance | tools/internal/storagecatalog ve contracts/storage owner infrastructure | Generic DB ownership manifests |
| Cutover authority | ops/cutoverauthority/adoption_catalog.go, contracts/cutover/v1/authority_catalog_digest.go | PostgreSQL object adoption ve authority inventory |
| Secret projector | ops/local/cmd/secret-projector/projection_catalog.go | Local secret projection manifest |
| Generic release | tools/localrelease/release_catalog.go | CatalogPackageSet alanı kaldırıldıktan sonra diğer release artifact'larının generic listesi korunabilir |

## 3. Benzer isimli ama farklı anlamdaki ayrımlar

- Billing catalog ile ForgeCRUD module Catalog aynı ürün kabiliyeti değildir. Billing offering/pricing/discount işlemleri kalır.
- IAM permission catalog ile Catalog modülünün catalog.read/catalog.install permission'ları aynı şey değildir. Yalnız Catalog module action mapping kaldırılır.
- REST Query field catalog ile Catalog package manifest'i farklıdır. Query field catalog kalır.
- PostgreSQL pg_catalog bir Harpy modülü değildir; DB engine metadata'sıdır.
- Generic async ProducerCatalog/ConsumerCatalog/Job Catalog veri tipleri kalır; yalnız Catalog'a ait producer/consumer/job kayıtları çıkarılır.
- System Resource Layout'ın catalog/query/presentation katmanı kalır; ForgeCRUD package menu versioning çıkarılır.
- Migration/storage catalog framework'leri kalır; yalnız Catalog owner/stream registration ve Catalog-owned SQL source çıkarılır.

## 4. Özellikle körlemesine silinmeyecek dosyalar

Aşağıdaki örnekler isim benzerliği nedeniyle korunmalıdır; call graph başka bir sonuca varırsa gerekçe ve test kanıtı olmadan silinmez:

- contracts/events/v1/catalog.go ve producer_catalog.go — generic event registry type'ları.
- contracts/jobs/v1/catalog.go — generic job registry type'ı.
- core/internal/platform/async/event_catalogs.go, release_catalogs.go, write_catalogs.go — generic async kayıt sistemi.
- core/internal/platform/migration/catalog.go, catalog_files.go, catalog_scope.go, tenant_catalog.go — generic migration catalog.
- tools/internal/migrationstreamcatalog/ ve tools/internal/storagecatalog/ — generic manifest loaders/validators.
- core/internal/platform/fieldcatalog/ ve REST Query field_catalog dosyaları — field schema.
- core/internal/modules/billing/application/restquery/reports/catalog.go ve Billing catalogquery/discountcatalog packages — commercial catalog/reporting.
- core/internal/modules/iam permission request/resource/action catalog ve identity/system resource catalog dosyaları — IAM native owner.
- core/internal/orchestration/systemresources/catalog.go ve core/internal/modules/system_resource_layout/application/catalog.go — system resource listing/presentation.
- ops/cutoverauthority/adoption_catalog.go ve contracts/cutover/v1/authority_catalog_digest.go — DB object authority migration.
- ops/local/cmd/secret-projector/projection_catalog.go — secret projection.
- tools/localrelease/release_catalog.go — generic artifact collection; yalnız Catalog package-set field'i çıkabilir.

## 5. Karar kuralı

Bir dosya adı veya sembolünde catalog geçmesi silme kararı değildir. Silme kararı için şu kanıtlar gerekir:

1. Dosyanın ForgeCRUD package registry / install / component / menu / demo / adoption özelliğinin parçası olması.
2. Call graph ve gerçek consumer'ların Catalog removal scope'unda olduğunun gösterilmesi.
3. Başka bir native Harpy owner'ın bu API/table/contract'ı kullanmadığının doğrulanması.
4. Gerekli ise data owner ve migration sahibi tarafından disposition onayı verilmesi.
5. Silme sonrası korunan generic owner için regression test bulunması.

İki anlama da gelebilen bir hit varsa varsayılan karar **koru ve incele**; toplu silme yapma.
