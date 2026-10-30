# 02 — Hedef sınırlar ve veri kararı

## 1. Hedef mimari

Harpy'nin hedef mimarisinde ForgeCRUD'a ait package/catalog installation katmanı bulunmayacak. Harpy, kendi business capability modüllerini ve bunları çalıştıran runtime süreçlerini doğrudan composition ile kuran backend platformu olarak kalacak.

Harpy'de kalacak modül ve capability sahiplikleri, kaynak kodda gerçekten bulunan native modüllerdir: IAM, Tenancy, Billing, Work, Documents, Notifications, Analytics, Audit, Usage, Platform Authority, Platform Ops, Support ve System Resource Layout. Bunların modül sınırları Catalog tarafından kurulmuş bir package dependency grafiğine bağlı olmayacak.

**ForgeCRUD modülleri Harpy runtime'ının kurulum birimi değildir.** HR, Procurement, Sales CRM, Accounting, Inventory, Quality, Production, Maintenance ve Project paketlerinin manifest, signature, dependency, tenant install, upgrade, menu ve demo seed akışları ForgeCRUD tarafında kalmalıdır. Bu task onları Harpy içinde yeniden modellemez.

## 2. Çıkarılacak davranış sınırı

Aşağıdaki ürün yetenekleri birlikte çıkarılmalıdır; yalnız registry API'sini kaldırıp arka taraftaki state machine veya owner adapter'larını bırakmak tamamlama değildir:

1. Signed immutable package/version registry, package manifest doğrulaması ve release publication.
2. Package dependency resolution ve exact version install plan.
3. Tenant package installation state ve install/upgrade/enable/disable/restore/retire/purge operation lifecycle.
4. Component binding, owner lookup/directory, component stage/apply/verify/activate/decommission/purge akışları.
5. Package kaynaklı menu definition/version/publication ve tenant menu override.
6. Catalog'a ait demo handoff, demo materialization, demo receipt ve demo removal lifecycle.
7. Catalog V1 adoption, legacy module/forms/menu storage inventory, compatibility adapter, cutover hit tracking ve retirement gate'leri.
8. Bu yeteneklere özel HTTP endpoints, IAM Catalog permission/action/capability'leri, worker processor'ları, job/event producer-consumer policy'leri.
9. Signed Catalog package release artifact'ı, public-key keyring, deploy evidence doğrulama ve setup içindeki local demo/package üretimi.
10. Catalog'a ait platform/tenant migration'ları, storage owner kaydı, SQL routines, tables, grants ve generated contract'lar.

Bu maddelerdeki davranışları başka bir adla taşıyan yeni bir package loader, registry, module install engine, component marketplace veya menu package system kurulmayacak.

## 3. Harpy'de kalacak benzer isimli kavramlar

Aşağıdaki yüzeyler farklı owner ve farklı işlevlere sahiptir. Catalog çıkarılırken kaldırılmayacak, yeniden adlandırılmayacak veya toplu regex ile değiştirilmeyecek:

- PostgreSQL sistem schema'sı pg_catalog ve tenant transaction search_path kuralları.
- REST Query field catalog, relation catalog, field profile, query operator/aggregate validation.
- Billing offering/catalogquery/discountcatalog ve billing commercial katalogları.
- IAM resource/action/permission catalog'u, permission request resource/action katalogları, RBAC, system resource ve profile katalogları.
- System Resource Layout'ın kendi revisioned resource/presentation catalog'u.
- Work storagecatalog ve REST Query storage/field mapping.
- Genel async producer catalog, consumer catalog, job kind catalog, release catalog abstractions.
- Genel migration catalog, migration stream catalog, storage owner catalog, PostgreSQL object/authority catalog, cutover adoption catalog.
- Local secret projector ve runtime config için kullanılan generic projection catalog'ları.

Bunların kesin path listesi 05 numaralı koruma matrisindedir. Bir path aynı anda hem Catalog ürün özelliği hem generic katalog sözcüğü içeriyorsa package import, owner ve call graph ile karar verilir.

## 4. IAM ve Analytics için hedef

### IAM

IAM'in Catalog package provisioning bağımlılığı çıkarılır:

- Catalog endpoint permission resolver ve Catalog tenant-release capability değerleri.
- Catalog owner olarak IAM_TEMPLATE component apply/projection adapter'ı.
- Catalog package payload/key contract'ına bağlı IAM catalogowner adapter'ı ve yalnız bu akışa ait migration/routine'ler.
- Catalog component lifecycle operation authority'leri.

IAM'in authentication, identity, session, organization, RBAC, resource/action/permission, permission request, authorization, system resource ve profile davranışları aynen kalır. IAM kendi permission catalog'unu korur; ForgeCRUD package kataloğunu IAM permission catalog'u ile karıştırma.

### Analytics

Analytics'in Catalog component provisioning bağımlılığı çıkarılır:

- catalogcomponentowner ve cataloglegacyowner application adaptörleri.
- catalogowner ve cataloglifecycle PostgreSQL adapter'ları ve bu adapter'ların SQL/migration'ları.
- Catalog payload validation ve owner/projection contract tüketimi.
- Bu akışın purge/retire/restore ve compatibility entegrasyon testleri.

Analytics'in native query, dashboard, export, publication, ACL, resource entitlement, REST Query ve kendi endpoint/data owner davranışları kalır. Catalog tarafından yüklenen bir dashboard/endpoint paketi artık Harpy tarafından install edilmez; Analytics capability'lerinin kendi native kurulum ve çalışma yolu korunur.

## 5. Veri sınıflandırması ve karar kapısı

Catalog'u çıkarmak, Catalog'a ait verilerin otomatik olarak silinebileceği anlamına gelmez. Önce en az aşağıdaki state türlerinin varlığı, sahibi, tenant kapsamı, satır sayısı, referansları ve retention gereksinimi çıkarılmalıdır:

| Veri sınıfı | Örnek state | Gereken işlem |
| --- | --- | --- |
| Package registry | Package, version, manifest, signature, owner payload | Export/retention/drop kararı |
| Tenant installation | Installed package/version, status, revision | Tenant bazlı envanter ve disposition |
| Dependency plan | Resolved versions, dependency blockers, plan digest | Export veya kontrollü kaldırma |
| Lifecycle operation | Install/upgrade/disable/purge operation, step, receipt, retry | Terminal/in-flight ayrımı ve disposition |
| Component binding | IAM/Analytics owner bindings, component state, owner receipts | Owner modül data impact analizi |
| Menu | Published version, items, tenant override, active pointer | Kullanıcı görünürlüğü ve audit etkisi |
| Demo state | Demo form rows, receipts, removal plans | Demo-only olduğunu kanıtla; kullanıcı verisiyle karıştırma |
| Adoption/cutover | V1 inventory, hit, acknowledgement, retirement gate | Cutover tamamlanma ve audit kanıtı |
| Security state | Catalog grants, permission mappings, step-up evidence | IAM cleanup ve audit/retention kararı |

Catalog owner sözleşmesinde görünen platform.catalog_package_owner_payloads, platform.catalog_package_versions, platform.catalog_packages tabloları ile menu_configs, module_dependencies, module_forms, module_migration_runs, module_schema_versions ve modules legacy table inventory'si ayrı ayrı sorgulanmalıdır. Bu isimlerin manifestte bulunması, canlı DB'de hâlâ bulunduklarını veya boş olduklarını kanıtlamaz.

IAM ve Analytics Catalog integration tabloları Catalog-owned tablolardan farklıdır. Bunların cleanup'ı kendi owner modül migration'ı ve owner contract'ı üzerinden ele alınır; Catalog SQL'inin içinden başka modül tablosu DROP edilmez.

## 6. Veri saklama / silme seçimi

Destructive migration yazılmadan önce aşağıdaki kararlardan biri açıkça seçilmeli ve kullanıcı/owner tarafından onaylanmalıdır:

- **Retain/export:** Catalog geçmişi denetim veya customer data gereği tutulur. Export formatı, encryption, access owner, retention, restore testi ve artifact lifecycle tanımlanır. Harpy runtime'ı bu export'u çalıştırmaya bağımlı olmaz.
- **Transform:** İş açısından gerekli kayıtlar belirlenmiş native owner state'e aktarılır. Her mapping, referential integrity, permission ve rollback/forward-recovery kanıtı test edilir. ForgeCRUD package modelini Harpy'ye taşımak transform kabul edilmez.
- **Drop:** Yalnızca state'in gereksiz olduğu, hiçbir consumer'ın kullanmadığı ve retention/contract yükümlülüğü bulunmadığı kanıtlanırsa uygulanır. Onaylı kapsam, tenant count ve object inventory kaydedilir.

Karar verilmedikçe default davranış **veriyi silmemek ve destructive DB değişikliği yapmamaktır**. Kullanıcının local volume'ları veya gerçek DB'si hiçbir test/agent tarafından otomatik reset edilmez.

## 7. API ve dış consumer uyumluluğu

API routes kaldırılmadan önce sibling frontend repository'si, MCP/external clients, deployment tooling, operator scripts, webhook/event consumers ve integration tests taranmalıdır. Harpy repository'sinde consumer bulunmaması, dış consumer bulunmadığını kanıtlamaz.

Catalog endpoint'lerini kullanan consumer varsa:
- Owner ve kullanım akışı kaydedilir.
- Consumer release/disable sırası planlanır.
- Gerekli data export veya yeni owner API geçişi kararlaştırılır.
- Eski endpoint için sessiz no-op veya sahte success bırakılmaz.
- Cutover sonrasında endpoint ve permission tamamen kaldırılır.

## 8. Migration yaklaşımı

Catalog'un 54 kendi migration'ı ile IAM/Analytics tarafında Catalog için eklenmiş migration'lar, mevcut migration/baseline sisteminin gerçek sahiplik kurallarıyla değerlendirilir. Generic migration catalog altyapısı çıkarılmaz.

Fresh install ve mevcut kurulum için davranış ayrı doğrulanır. Eğer migration-baseline-rework işi hâlâ aktifse, onun shared migration/build/manifest dosyalarıyla çakışan işler aynı anda yürütülmez. Baseline artefact, owner manifest veya generated migration aggregate'i tek owner geçişinde güncellenir.

Catalog object removal, yalnızca bağımlılıkları kaldırılmış ve veri kararı verilmiş nesneleri kapsar. DROP sırası view/function/trigger/grant/table/schema dependency graph'ından çıkarılır; tahmine dayalı cascade kullanılmaz.

## 9. Karar kapısı

Faz 0 bitmeden şu soruların cevabı yazılı değilse destructive implementation başlamaz:

- Hangi production/local tenant'larda Catalog state var?
- Catalog API'lerini hangi frontend veya dış consumer'lar çağırıyor?
- Catalog install/upgrade operation'ları arasında in-flight veya retry bekleyen iş var mı?
- IAM/Analytics component payload ve owner data için retain/export/drop kararı nedir?
- Menu override, publication, audit ve permission history için retention gereksinimi var mı?
- Mevcut fresh-install/baseline/migration akışı hangi owner tarafından güncellenecek?
- Eski deploy artifact'ları ve keyring rotation/revocation nasıl sonlandırılacak?
