# 01 — Mevcut kod envanteri ve cerrahi kaldırma haritası

## 1. Envanterin kapsamı

Bu liste başlangıç haritasıdır; tüm kaldırılacak dosyaların kapalı listesi değildir. Uygulama başlamadan önce güncel repository üzerinde path, import, symbol, SQL object, embed, manifest ve runtime registration taraması tekrarlanmalıdır.

30 Ekim 2026 tarihli main ağacı ölçümünde core/internal/modules/catalog/ altında 646 tracked file ve kendi migration ağacında 54 SQL dosyası bulunuyordu. Modül yalnız HTTP handler'larından ibaret değildir; application, domain, postgres, HTTP, adoption, backup, worker, demo ve lifecycle alanlarını kapsar.

## 2. Catalog modülünün kendi yüzeyi

Ana kök: **core/internal/modules/catalog/**

| Alan | Mevcut sorumluluk |
| --- | --- |
| domain/packagecontract/ | Signed package manifest, identity, version, dependency ve validation |
| application/registry/ | Package registry, package list/get/version resolution |
| application/dependency/ | Required/optional dependency resolution |
| application/lifecycle/ | Install, upgrade, enable, disable, restore, retire, purge, operation state |
| application/menu/ | Menu definition, version, publication ve tenant override |
| application/adoption/ | Eski V1 module/catalog storage adoption ve cutover |
| application/demohandoff/, demomaterialization/ | Demo payload, receipt, materialization ve handoff |
| application/releasepublication/ | Package release publication ve payload binding |
| adapters/http/ | Registry, dependency, installation, lifecycle, operation, menu ve demo endpoint'leri |
| adapters/postgres/ | Registry, lifecycle, menu, adoption, owner ve backup persistence |
| migrations/platform/, migrations/tenant/ | Catalog'a ait platform ve tenant DB nesneleri |
| Root async/worker dosyaları | Durable lifecycle jobs, event/outbox, retry/resume ve worker processor |
| Root HTTP/runtime dosyaları | HTTP facade, scope, authorization, runtime ve route binding |

Modülün içindeki docs/freeze-handoff.md, migration_documents.go, phase/gate/migration testleri ve SQL go:embed kullanımları da envantere dahildir. Yalnızca Go paket klasörlerini silip dışarıdaki embed veya manifest referanslarını bırakmak kabul edilmez.

## 3. Runtime composition bağlantıları

- core/internal/composition/api/build_catalog_http.go: Catalog API runtime'ını kurar; IAM access verifier, owner directory, package keyring, transaction coordinator ve Catalog handler'ı bağlar.
- core/internal/composition/api/catalog_access.go: Catalog endpoint'lerinin IAM permission/authority erişimini doğrular.
- core/internal/composition/worker/build_catalog_job.go: Catalog lifecycle job processor'ını, owner directory ve package keyring ile kurar.
- İlgili composition testleri: build_catalog_http_test.go, catalog_access_test.go, build_catalog_job_test.go ve catalog_demo_*_integration_test.go.
- API route assembly ve readiness kaynaklarında Catalog runtime'ın eklenmesi/ready sayılması bulunur. Sadece builder dosyasını silmek yeterli değildir.

Taranacak kaynaklar: core/internal/composition/api/, core/internal/composition/worker/, contracts/http/, runtime readiness, operation/authority aggregates ve worker processor registration.

## 4. Catalog'a bağlı owner modülleri

### IAM

Catalog'a özel bağlantılar arasında:

- core/internal/modules/iam/catalog_access_permission.go: catalog.read/install/upgrade/disable/enable/restore/retire/purge/demo.manage/menu.manage permission ve operation eşlemesi.
- core/internal/modules/iam/domain/sessioncapability/catalog.go: tenant release capability listesinde Catalog lifecycle permission'ları.
- core/internal/modules/iam/catalog_component_owner.go ve adapters/postgres/catalogowner/: IAM_TEMPLATE component provisioning / Catalog owner API.
- core/internal/modules/iam/api/catalogpayload/ ve ilgili package validation.
- iam_catalog_component_*, iam_catalog_projection_*, iam_catalog_version_*, iam_catalog_operational_* gibi Catalog component/package lifecycle migration ve SQL'leri.

**Korunacak IAM yüzeyleri:** IAM resource/action/permission catalog'u, RBAC, permission request resource/action catalog'ları, authentication, identity, system resource ve profile işlemleri. İsimde catalog geçmesi tek başına silme gerekçesi değildir.

### Analytics

Catalog'a özel bağlantılar arasında:

- core/internal/modules/analytics/catalog_component_owner.go.
- application/catalogcomponentowner/ ve application/cataloglegacyowner/.
- adapters/postgres/catalogowner/ ve adapters/postgres/cataloglifecycle/.
- api/catalogpayload/.
- analytics_catalog_projection_*, component apply/blocker, legacy bridge ve purge lifecycle migration'ları.
- Catalog owner/projection kullanan PostgreSQL integration fixture ve testleri.
- contracts/operations/analytics_application_catalogcomponentowner.json ve ...cataloglegacyowner.json.

**Korunacak Analytics yüzeyleri:** Native saved query, query execution, dashboard, export, publication, access control, REST Query ve Analytics'in kendi kaynak verisi/endpoint'leri. Catalog üzerinden gelen package provisioning'in kaldırılması Analytics ürün davranışlarını toptan kaldırma gerekçesi değildir.

## 5. Shared capability ve wire contract bağlantıları

| Yol | Catalog'a özel içerik |
| --- | --- |
| core/internal/capabilities/catalogcomponent/ | Component owner persistent state ve apply/blocker/read interface'leri |
| core/internal/capabilities/catalogprojection/ | Package/component projection ve owner provisioning |
| contracts/catalogcomponent/v1,v2/ | Component owner contract'ları |
| contracts/catalogpackage/v1/ | Signed package set, manifest ve owner payload contract'ları |
| contracts/http/catalog-v2.* | Catalog registry/lifecycle/menu endpoint schema ve paths |
| contracts/authority/catalog_*.json | Catalog authority tanımları |
| contracts/operations/catalog_*.json | Catalog operation/source/test registration |
| contracts/asyncmanifest/v1/catalog_*.go | Catalog event producer/consumer policy |
| contracts/events/v1/catalog-*.schema.json | Catalog installation/menu event schema'ları |
| contracts/jobs/v1/ | Catalog'a özel job kind/payload policy registration'ı |

Aggregate dosyaları (ör. authority.yaml, operations.yaml, OpenAPI, async manifests) başka modülleri de içerir. Bunlar Catalog dosyalarıyla aynı şey değildir; yalnız Catalog include/registration satırları çıkarılır.

## 6. Release, setup ve attestation bağlantıları

- cmd/setup/release_catalog_document.go, release_catalog_manifest.go ve testleri: local signed Catalog package/demo payload üretimi.
- contracts/catalog/firstparty-parity.v1.json ve inventory: ForgeCRUD first-party module packages için parity manifesti.
- internal/attestation/verify_catalog_package_set.go, verified_catalog_package_set.go: release package set doğrulaması.
- core/internal/platform/deployevidence/verify_catalog_packages.go: migrator deploy evidence içinde signed Catalog package set doğrulaması.
- core/internal/platform/config/catalog_package_keyring.go, runtime schema ve fixture/testler: HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING configuration surface'i.
- tools/localrelease/catalog_package_*.go, catalog_reference.go, catalog_demo_document.go, existing_output_catalog.go ve release artifact set wiring'i.
- tools/localconfig/ altında migrator Catalog package artifact/keyring projection ve validation dosyaları.

tools/localrelease/release_catalog.go gibi isimleri generic olan dosyalar release manifest'inin genel taşıyıcısı olabilir. Dosya adını değil field/call graph sahipliğini takip et; migration, roles, privileges, extensions, ownership ve bootstrap artifact'larını kaybetme.

## 7. Database ownership ve migration kaynakları

- Catalog module platform ve tenant migration dizinleri: core/internal/modules/catalog/migrations/**.
- Catalog SQL adapters: adapters/postgres/**/sql/**.
- Storage owner: contracts/storage/owners/catalog.json.
- Migration stream: contracts/storage/migration-streams/catalog.json.
- Authority/operation/event contract'ları ve IAM/Analytics owner migration'ları.
- Catalog owner tanımında görünen platform tables: platform.catalog_package_owner_payloads, platform.catalog_package_versions, platform.catalog_packages.
- Owner dosyasındaki legacy table inventory: menu_configs, module_dependencies, module_forms, module_migration_runs, module_schema_versions, modules.

Legacy tablo adlarının owner manifestinde yer alması, tabloların halen aktif olduğu veya boş olduğu anlamına gelmez. Her nesnenin gerçek DB varlığı, owner/grant/ACL, trigger/function ilişkisi, satır sayısı, referansları ve yasal/data retention gereksinimi uygulama öncesi doğrulanmalıdır.

## 8. Arama prosedürü

İlk tarama için:

    rg -n -i "core/internal/modules/catalog|modules/catalog|catalogpackage|catalogcomponent|catalogprojection|catalog.installation|catalog.menu|catalog.demo|catalog.read|catalog.install|catalog.upgrade|CatalogPackageKeyring|CATALOG_PACKAGE_PUBLIC_KEYRING" .
    rg -n "buildCatalogHTTP|buildCatalogJobProcessor|NewPlatformHTTPRuntime|NewPlatformJobProcessor|CatalogComponentOwner|ComponentOwnerDirectory" core cmd
    rg -n "catalog_" core/internal/modules contracts
    rg -n "catalog" contracts/authority.yaml contracts/operations.yaml contracts/storage

Sonra her hit için: production import, test-only import, generated reference, SQL object, migration, go:embed, JSON/YAML include, config env, OpenAPI, worker manifest ve external consumer şeklinde sınıflandırma yap. Tüm catalog kelimesini sıfırlamak hedef değildir; izin verilen generic kullanımlar 05 numaralı matriste listelenir.

## 9. Envanter çıkış kanıtı

- Her Catalog-specific symbol ve path'in sahibi belli.
- Her dış consumer ve endpoint kullanıcısı tespit edilmiş.
- IAM ve Analytics için Catalog dışı kalan native işlevler ayrı listelenmiş.
- Her Catalog-owned DB object ve runtime-generated data için veri kararı gereksinimi kaydedilmiş.
- Tüm generated/aggregate call site'ları ve silme sırası tespit edilmiş.
- Belirsiz bir ownership veya consumer kalmışsa kod silme fazına geçilmemiş.
