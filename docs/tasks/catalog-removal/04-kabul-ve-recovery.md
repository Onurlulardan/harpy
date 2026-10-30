# 04 — Kabul kriterleri, testler ve recovery

## 1. Definition of Done

Bu çalışma ancak aşağıdakilerin tamamı sağlandığında kapanır:

- [ ] Harpy'de ForgeCRUD package registry, signed package set, dependency resolver veya tenant module install lifecycle bulunmuyor.
- [ ] Catalog HTTP route, request/response schema, OpenAPI entry, IAM Catalog permission mapping ve authority kaydı kalmıyor.
- [ ] API composition Catalog HTTP runtime/readiness oluşturmaz; Worker Catalog job processor veya Catalog event consumer/producer kaydetmez.
- [ ] IAM ve Analytics Catalog component owner/projection entegrasyonlarından bağımsız derlenir ve çalışır.
- [ ] IAM authentication, identity, RBAC, resource/action/permission catalog ve permission request işlevleri korunur.
- [ ] Analytics native query, dashboard, publication, export, ACL ve endpoint işlevleri korunur.
- [ ] Setup/localrelease/deploy evidence Catalog package üretmez, imzalamaz, keyring istemez veya package set doğrulamaz.
- [ ] Catalog-specific tenant/platform migrations, SQL routines, storage owner, authority, operation, event ve async registrations temizlenmiştir.
- [ ] Catalog-owned data için retain/export/drop kararı, onay ve doğrulama kanıtı vardır.
- [ ] Fresh platform install ve tenant provisioning Catalog olmadan başarılıdır.
- [ ] Migration/baseline akışı Catalog-specific schema owner olmadan doğrulanmıştır; generic migration engine değişmeden kalır.
- [ ] Catalog module tree, capability packages, package contracts, tests, embeds, fixtures ve stale generated artifacts kaldırılmıştır.
- [ ] README ve architecture docs artık Catalog'u Harpy business module/capability olarak tanıtmamaktadır.
- [ ] Generic catalog kavramları koruma matrisiyle doğrulanmıştır; toplu catalog string removal yapılmamıştır.
- [ ] go test ./..., go vet ./..., fileguard, archguard, contract/OpenAPI checks, fresh-install integration ve ilgili E2E kontrolleri geçmiştir.

## 2. Statik referans ve paket kontrolleri

Aşağıdaki hedefler source tree ve generated çıktılarda bulunmamalıdır:

- core/internal/modules/catalog/
- core/internal/capabilities/catalogcomponent/
- core/internal/capabilities/catalogprojection/
- contracts/catalogpackage/
- contracts/catalogcomponent/
- docs/architecture/modules/catalog.md
- skills/catalog/ (final docs cleanup fazından sonra)
- build_catalog_http.go ve build_catalog_job.go
- Catalog-specific registry/lifecycle/menu/demo HTTP routes
- Catalog-specific authority/operation/event/job/async/storage owner registrations
- HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING ve CatalogPackageKeyring runtime config
- Catalog release package, first-party parity package set ve deploy verification fields
- IAM/Analytics Catalog owner adapter ve Catalog-specific migration/import references

Önerilen targeted taramalar:

    rg -n "core/internal/modules/catalog|modules/catalog" .
    rg -n "catalogpackage|catalogcomponent|catalogprojection" .
    rg -n "catalog.installation|catalog.menu|catalog.demo|catalog.package" .
    rg -n "CatalogPackageKeyring|CATALOG_PACKAGE_PUBLIC_KEYRING|buildCatalogHTTP|buildCatalogJobProcessor" .
    rg -n "catalog_" core/internal/modules/iam core/internal/modules/analytics contracts

Her residual hit açıklanmalıdır. Billing, IAM permission, field catalog, async registry, migration/storage catalog veya pg_catalog gibi korunacak bir kavrama aitse preservation matrix'teki allowlist'e eklenip owner/scope gerekçesiyle raporlanır. Bu taramaların amacı her catalog kelimesini yok etmek değildir.

## 3. Go ve architecture testleri

Repository root:

    go test ./...
    go vet ./...
    go run ./tools/fileguard
    go run ./tools/archguard
    go mod tidy
    git diff --exit-code -- go.mod go.sum

go mod tidy sonrası dependency değişiklikleri bekleniyorsa diff gerekçelendirilir. Generic capability veya başka modül dependency'si yanlışlıkla kaldırılmamalıdır.

Targeted test alanları:
- IAM module ve PostgreSQL integration tests.
- Analytics module, query/dashboard/publication ve PostgreSQL integration tests.
- API composition, public route/OpenAPI aggregate ve management readiness tests.
- Worker composition, event producer/consumer policy, job manifest ve processor registry tests.
- Setup CLI, localrelease, deploy evidence, attestation ve runtime config tests.
- Storage owner/migration stream/authority/operation aggregate validation.
- REST Query, Billing catalogquery/discountcatalog, IAM permission catalog, System Resource Layout, Work storage catalog ve generic async/migration catalog regression tests.

## 4. PostgreSQL ve migration test matrisi

| Senaryo | Beklenen |
| --- | --- |
| Fresh PostgreSQL 18 platform install | Catalog owner/table/routine gerektirmeden tamamlanır |
| Fresh tenant provisioning | Tenant schema ve IAM/Analytics native işlevleri Catalog'suz hazırlanır |
| Birden fazla tenant | Tenant scope, grants ve schema isolation etkilenmez |
| Catalog state mevcut DB | Faz 0'da kararlaştırılan retain/export/drop prosedürü uygular |
| Export seçeneği | Export digest, encryption, row counts ve restore testi doğrulanır |
| Drop seçeneği | Yalnız onaylı Catalog objects kaldırılır; owner dışı DB objesi etkilenmez |
| Retry/in-flight operation | In-flight Catalog operation için kabul edilen disposition deterministik ve denetlenebilir |
| Backup/restore | Catalog removal sonrası kalan Harpy modüllerinin backup/restore işlemi tamamlanır |
| İkinci migration up | Yeni migration runner/baseline mimarisine göre no-op veya beklenen idempotent sonuç verir |
| Concurrent migrator | Aynı migration scope'unda duplicate DDL veya yarım cleanup oluşmaz |

Migration-baseline-rework ile aynı zamanda yürütülüyorsa exact commands ve fixture strategy o task'ın güncel kabul kriterleriyle koordine edilir. Kullanıcı volume'u veya gerçek DB, test ortamı olarak kullanılmaz.

## 5. API ve async kabul testleri

- Eski Catalog path'leri için route registry 404/route-not-registered sonucu üretir; geride sahte 200/no-op endpoint kalmaz.
- OpenAPI aggregate ve runtime path tests Catalog endpoint'i içermez.
- IAM Catalog capability/permission değerleri tenant release allowlist'inde bulunmaz; diğer IAM permission değerleri değişmez.
- Event producer/consumer catalog'larında catalog.installation.* ve catalog.menu.* event policy'leri yoktur.
- Job kind catalog'unda Catalog lifecycle processor'a ait job türü yoktur.
- Worker unknown Catalog event/job ile fail-closed davranır; unknown event'i sessizce yutmaz.
- API readiness, Worker readiness ve setup/operator initialize Catalog runtime veya keyring aramaz.
- Billing commercial catalog API'leri, REST Query ve diğer native Harpy endpoint'leri contract testleriyle doğrulanır.

## 6. Veri ve güvenlik regresyonları

- Export edilen package/installation/menu/owner state kayıtlarının source count ve digest doğrulaması yapılır.
- IAM/Analytics Catalog-specific owner tables temizlenirken native owner tabloları, grants ve routines etkilenmez.
- Permission mapping kaldırılması IAM RBAC create/update, permission request review/approve ve tenant isolation davranışını bozmaz.
- Menu visibility'nin authorization olmadığına dair genel güvenlik kuralı ve native System Resource Layout sunumu korunur.
- Secret, signing key, credential, tenant data veya raw document content test loglarına yazılmaz.
- Catalog keyring env kaldırılırken release signing, license key, recovery key ve diğer keyring'ler korunur.

## 7. Geri alma ve recovery

- Git geçmişi yeniden yazılmaz; Catalog extraction commit'leri normal source history'de kalır.
- DB data export gerekiyorsa destructive migration öncesi ayrı encrypted artifact üret ve restore edilebilirliğini test et.
- Drop migration uygulanmışsa yalnız kod rollback'ine güvenme; forward recovery veya doğrulanmış data restore prosedürünü yaz.
- Eski endpoint/event/job consumer'a geri dönme planı gerekiyorsa owner/client release sırasıyla tanımla. Sahte compatibility shim bırakma.
- Başarısız fazda kullanıcı local named volume'ları, production DB veya secret volume'ları otomatik silinmez.
- task local:destroy, docker volume prune, DB reset veya force drop komutları agent tarafından otomatik çalıştırılmaz.

## 8. Final rapor şablonu

- Değişen/silinen dosya ve package listesi; her grubun sahibi ve sebebi.
- Kaldırılan endpoint, permission, authority, event, job, keyring ve migration listesi.
- Korunan generic catalog kavramları ve bunların regression testleri.
- IAM ve Analytics için targeted test + PostgreSQL integration sonuçları.
- Fresh platform/tenant install ve API/worker readiness sonuçları.
- Catalog data disposition: retain/export/transform/drop, onaylayan owner ve doğrulama kanıtı.
- go test, vet, fileguard, archguard, OpenAPI, build ve E2E komutları, exit code'lar.
- Kalan her Catalog referansı ve neden kaldırılmadığı.
- Destructive local/production action uygulanıp uygulanmadığı (varsayılan: hayır).
