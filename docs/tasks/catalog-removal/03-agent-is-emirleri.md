# 03 — Agent iş emirleri ve faz sırası

## Genel uygulama protokolü

Bu çalışma cross-module, public API, tenant data ve migration etkisi taşır. Kaynak değiştirmeden önce şu skill'leri eksiksiz oku: code-organization, modular-architecture, impact-analysis, api-contract, data-tenancy, events-jobs, testing-quality ve hâlâ mevcutsa skills/catalog/SKILL.md. Catalog skill'inde anılan ADR-0015, ADR-0026, ADR-0032, ADR-0117 ve phase14 platform capability matrix'i de okunmalı; ilgili ADR'ler authoritative domain kararlarını içerir.

Tek bir integration branch üzerinde ilerle. Fazlar kabul edilmeden sonraki fazın destructive işi başlatılmaz. Shared aggregate ve generated dosyalarda tek owner kullanılır. Paralel agent'lara aynı path yazma yetkisi verilmez. Her faz sonunda diff, test ve kalan reference raporu gerekir.

## Faz 0 — Salt okunur sahiplik ve consumer audit

**Sahip:** Read-only audit agent. **Yazılabilir path yok.**

1. 01 numaralı envanteri güncel tree ile doğrula; her eklenen/değişen path'i raporla.
2. Catalog module, API, worker, IAM, Analytics, contracts, migrations, setup/release, attestation, config, localconfig, ops ve docs call graph'ını tara.
3. Public HTTP endpoint method/path, IAM permission, authority, job kind, event producer/consumer, keyring env ve signed artifact envanterini çıkar.
4. Sibling frontend, MCP/external clients, operator scripts ve deploy automation'da Catalog tüketimini araştır.
5. Platform/tenant DB nesneleri ile tenant state sınıflarını data kararı tablosuna bağla.
6. Generic catalog kullanımını 05 matrisindeki koruma listesiyle eşleştir; yeni generic hit'leri listeye ekle.
7. Açık karar ve dış blocker'ları çıkar.

**Çıkış:** Owner/consumer listesi, endpoint ve contract envanteri, data inventory, test baseline ve onay bekleyen data disposition. Belirsizlik varsa silme başlamaz.

## Faz 1 — IAM ve Analytics owner bağımlılıklarının çözülmesi

**Sahiplik:** IAM modül path'leri ve Analytics modül path'leri ayrı agent'larda. Catalog modülüne ve shared aggregate'lere yazılmaz.

### IAM agent

- Catalog access permission resolver ve yalnız Catalog lifecycle'a ait tenant capability kayıtlarını kaldır.
- IAM Catalog component owner, IAM_TEMPLATE provisioning, catalogpayload ve catalogowner adapter bağımlılıklarını ayır.
- Catalog component/projection için oluşturulmuş IAM-specific SQL routines/tables/migrations ve testleri envantere göre temizle.
- IAM auth, identity, RBAC, permission request resource/action catalog, system resource ve profile davranışlarını koru.
- IAM module composition ve test fixture'larının Catalog owner beklemediğini kanıtla.

### Analytics agent

- Analytics Catalog component owner ve legacy owner application adaptörlerini kaldır.
- catalogowner/cataloglifecycle PostgreSQL adapter, SQL ve Catalog-specific migration/testleri temizle.
- catalogpayload ve contracts/catalogcomponent/catalogpackage import'larını kaldır.
- Native Analytics query/dashboard/export/publication/ACL ve endpoint davranışlarını koru.
- Analytics install/package lifecycle'a bağlı olmadığını ve native runtime'ın bağımsız çalıştığını test et.

**Faz çıkışı:** IAM ve Analytics kendi testleriyle Catalog olmadan kurulabilir; cross-module SQL/import kalmamıştır. Catalog capability ve contract paketleri bu faz sonunda hâlâ başka consumer kullanıyorsa silinmez.

## Faz 2 — API, permission ve worker runtime bağlantısının sökülmesi

**Sahiplik:** core/internal/composition/**, Catalog-specific HTTP/async contracts ve API aggregate owner'ı.

- API composition'dan build_catalog_http.go, Catalog access verifier, owner directory ve Catalog readiness/route registration'ı çıkar.
- Public route registry, OpenAPI paths/schemas, Catalog endpoint decoder/presenter ve integration fixtures kaldırılır.
- IAM Catalog permission/authority kontrollerinin endpoint erişimi için artık çağrılmadığını doğrula.
- Worker composition'dan Catalog job processor, owner directory, package keyring ve Catalog event producer/consumer/job registration'ını çıkar.
- Catalog installation lifecycle ve menu event schema'larını kaldır; generic producer/consumer catalog abstractions korunur.
- API liveness/readiness, worker startup ve manifest compatibility testleri Catalog registration olmadan geçmelidir.
- External consumer bulunduysa faz 0'da kararlaştırılan cutover sırasına göre consumer önce çıkarılır veya yeni owner'a geçirilir.

**Faz çıkışı:** API ve Worker Catalog package kurmadan ve Catalog runtime olmadan başlar; eski Catalog endpoint, job veya event registration'ı yoktur.

## Faz 3 — Signed release, setup ve deploy evidence temizliği

**Sahiplik:** cmd/setup/**, internal/attestation/**, core/internal/platform/deployevidence/**, core/internal/platform/config/**, tools/localconfig/**, tools/localrelease/**.

- Setup'ın local Catalog package, project demo payload, package manifest ve Catalog signing flow'unu kaldır.
- releaseKeys ve runtime config içindeki Catalog public/private key alanlarını kaldır; HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING ve ilgili schema/fixture/testleri temizle.
- Migrator deploy evidence içinden Catalog package set verification'ı çıkar.
- Attestation tarafında Catalog package-set marshal/verify modelleri ve testlerini kaldır.
- Local release artifact builder içinden Catalog package-set artifact'ını çıkar.
- tools/localrelease/release_catalog.go generic release manifest/artefact taşıyıcısıysa dosyayı silme; yalnız Catalog package set alanını ve ona ait verification/wiring'i kaldır.
- Bootstrap, migration, extensions, roles, privileges, ownership, credentials, installation, license, recovery ve release-signing artefact'larını koru.
- Localconfig Catalog artifact set projection/validation/keyring references temizlenirken generic artifact catalog ve recovery artifact registry'lerini koru.

**Faz çıkışı:** Fresh setup/release pipeline signed ForgeCRUD Catalog package üretmez, keyring istemez ve migrator deploy evidence Catalog package doğrulamasına bağlı değildir.

## Faz 4 — DB owner, migration ve data disposition

**Sahiplik:** Catalog module migrations + IAM/Analytics owner migrations + storage/authority owner kayıtları; tek migration entegrasyon sahibi.

1. Faz 0 data inventory ve Faz 2 consumer çıkışı tamamlanmadan DROP migration yazma.
2. platform.catalog_package_owner_payloads, platform.catalog_package_versions, platform.catalog_packages ve tenant Catalog objects için canlı DB owner, grants, routines, triggers, data count ve references doğrula.
3. Legacy module_forms/modules/module_dependencies/menu_configs/module_schema_versions/module_migration_runs tablolarının gerçek kullanımını ayrı ayrı sorgula; bunların tamamı Catalog owner manifestindeki inventory kayıtlarıdır, canlıda hepsi vardır diye varsayma.
4. IAM ve Analytics Catalog-specific storage cleanup'ını kendi owner migration/contract sınırlarında yap. Başka owner'ın tablosuna Catalog migration'ından SQL yazma.
5. Retain/export kararı varsa export digest, encryption, access, restore testi ve retention policy'si tamamlanmadan source data drop etme.
6. Drop kararı varsa explicit dependency graph ve reverse dependency sırasıyla migration planla; CASCADE, schema reset veya global permission/authority reset kullanma.
7. Catalog module migration stream ve storage owner include kayıtlarını shared aggregate update fazına hazırla.
8. Fresh install, mevcut supported DB upgrade ve tenant başına cleanup senaryolarını izole PostgreSQL üzerinde test et.

**Faz çıkışı:** Catalog-owned schema lifecycle kaldırılmış; native owner şemaları ve grants doğru; data disposition kanıtları ve rollback/forward-recovery planı mevcut.

## Faz 5 — Catalog module ve shared capability/contracts silinmesi

**Başlangıç koşulu:** Faz 1–4 tamamlandı ve Catalog modülünün production import/consumer sayısı sıfır.

- core/internal/modules/catalog/** tamamen kaldırılır: domain, application, adapters, migrations, tests, embedded SQL, docs ve root runtime files.
- core/internal/capabilities/catalogcomponent/ ve catalogprojection/ yalnızca gerçek consumer kalmadığı kanıtlandıktan sonra silinir.
- contracts/catalogpackage/, contracts/catalogcomponent/, Catalog-specific HTTP, authority, operation, event schema, async/job policy ve first-party parity manifestleri kaldırılır.
- contracts/storage/owners/catalog.json ve migration-streams/catalog.json silinir; shared root include listeleri tek entegrasyon owner'ı tarafından güncellenir.
- IAM/Analytics migration/test tree'de Catalog-specific stale SQL, fixtures, generated snapshots ve embed references kalmamalı.
- skills/catalog/SKILL.md ve skills/INDEX.md içindeki Catalog routing, feature tamamen çıkarıldıktan sonra temizlenir.
- core/internal/platform/async, migration, fieldcatalog, storagecatalog, REST Query ve generic catalog modellerine dokunulmaz.

**Faz çıkışı:** Harpy'de Catalog package manager veya component provisioning için source package, migration owner, contract veya runtime registration kalmamıştır.

## Faz 6 — Aggregate, generated output ve dokümantasyon normalizasyonu

**Sahip:** Tek integration/documentation owner. Bütün module agent'ları kapandıktan sonra.

- contracts/authority.yaml, contracts/operations.yaml, HTTP/OpenAPI aggregate, async manifests, storage owner/migration stream ve permission catalog'ları güncelle.
- Generated contract/schema/output'ları kendi generator'ı ile yeniden üret; elle generated dosya düzeltme.
- fileguard/archguard allowlist, package inventory, operation test claims, migration/storage owner counts ve expected snapshots güncelle.
- README ve docs/architecture içinde ForgeCRUD Catalog business modülüne ait tüm tanımları kaldır; System Resource Layout, REST Query ve Billing catalog açıklamalarını koru.
- docs/architecture/modules/catalog.md ve Catalog-specific skill dosyası silinmiş olmalı. Bu docs/tasks/catalog-removal task set'i yürütme kaydı olarak kalır.
- Eski docs/harpy-starter gibi silinmiş bir dizine referans ekleme; tüm relative links doğrulanır.

**Faz çıkışı:** Aggregate'ler tek tutarlı state'te, dokümanlar kaynak kodla uyumlu ve Catalog ürün özelliği Harpy'de tanıtılmıyor.

## Faz 7 — Bağımsız doğrulama

Bu agent implementation tasarımını değiştirmez; test, diff ve kalan referansları denetler. Kabul kriterleri 04 numaralı belgede tanımlıdır. Başarısız kapı varsa ilgili owner fazı yeniden açılır.
