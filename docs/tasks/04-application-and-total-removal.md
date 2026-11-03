# 04 — APPLICATION ve SIFIR KALINTI

## 1. Amaç

Bu görev iki sonucu birlikte tamamlar:

1. APPLICATION süreçlerinin yalnız DATABASE FOUNDATION ve GOOSE başarılı olduktan sonra çalışmasını sağlamak.
2. Eski migration altyapısına ait her executable ve non-executable kalıntıyı repository'nin tamamından kaldırmak.

Bu bir refactor, compatibility, migration veya gradual cutover görevi değildir. Eski sistemin tamamı silinir; yeni dört aşamalı sistem bağımsız olarak çalışır.

## 2. APPLICATION başlangıç kontratı

APPLICATION, INITIALIZE akışının son aşamasıdır. API, worker ve diğer uygulama servisleri yalnız Foundation ve Goose başarıyla tamamlandıktan sonra başlatılır.

API/worker startup davranışı:

- Migration çalıştırmaz.
- Goose çağırmaz.
- Schema veya database oluşturmaz.
- Table/column/index/constraint/function/trigger oluşturmaz.
- Extension kurmaz.
- Role/grant/ownership değiştirmez.
- Eksik schema'yı onarmaya kalkışmaz.
- Migration version veya release/generation bilgisi kontrol etmez.
- Custom schema contract veya readiness receipt okumaz.

Uygulama readiness kontrolü yalnız process'in normal çalışma bağımlılıklarını doğrular: DB bağlantısı, gerekli runtime rolü, uygulamanın ihtiyaç duyduğu servisler ve temel bağlantı kontrolleri. DDL, schema repair veya migration yürütmek readiness kontrolünün parçası olamaz.

Normal uygulama iş kuralları, tenancy, authorization ve domain davranışları bu task kapsamında gereksiz yere yeniden tasarlanmaz. Ancak eski migration framework'üne bağlanan her Go import'u, interface'i, service call'u, config alanı ve startup hook'u kaldırılır.

## 3. Tenant provisioning sınırı

Yeni tenant schema'sı eski tenant state'ini içeri almaz. Yeni tenant boş schema olarak oluşturulur; Foundation namespace/owner/privilege altyapısını hazırlar; aynı Goose runner tenant migration source'unu uygular; Goose başarılı olduktan sonra application provisioning akışı tenant'ı aktif eder.

Bu işlem için aşağıdakiler oluşturulamaz:

- Tenant migration executor.
- Tenant migrator service/daemon/container/binary.
- Migration'a özel queue consumer veya worker.
- Migration run/attempt/claim/lease/checkpoint/receipt tablosu.
- Migration completion stored procedure.
- Tenant schema version registry.
- Platform migrator ile tenant migrator arasında özel RPC/HTTP kontratı.

Tenant oluşturma domain iş akışı için zaten var olan genel provisioning mekanizması korunabilir; ancak migration state'i, schema DDL'i veya Goose history'si bu genel mekanizmanın içine ikinci bir migration sistemi olarak taşınamaz.

Tenant schema oluşturma ve Goose çalıştırma yetkisi API/worker runtime credential'ına verilmez. Bu operasyonun güvenli çağrı sınırı dört aşamalı lifecycle ve mevcut deployment güvenlik modeli içinde kurulur; yeni deployable veya daemon ekleyerek çözülmez.

## 4. Go source için sıfır kalıntı şartı

Repository'deki bütün **.go** dosyaları taranır. Eski migration framework'üne ait hiçbir executable kod, type, interface, function, import, constant, variable, package veya test helper kalamaz.

Silinecek kod kapsamı:

- Eski migration domain model'leri.
- Migration port/interface/adapter/store implementasyonları.
- Custom runner ve executor'lar.
- Migration planner/selector/classifier/parser/validator/verifier/finalizer.
- Migration manifest/release/catalog/digest model'leri.
- Migration run/attempt/checkpoint/receipt/requirement akışları.
- Goose'u saran eski policy/state machine kodları.
- Platform migrator service/command/process.
- Tenant migration executor/service/worker/process.
- Eski initialization ve migration orchestrator bağlantıları.
- Migration-specific retry/reconcile/cleanup/fencing akışları.
- Eski SQL routine'lerini veya metadata tablolarını yöneten repository/store kodları.
- Eski sistemin generated Go contract/model dosyaları.
- Eski migration altyapısına özel test helper, fixture builder ve mock'lar.

Her silme işleminde import graph ve consumer zinciri takip edilir. Sadece package klasörünü kaldırmak yeterli değildir; consumer'ların tamamı güncellenir. Kullanılmayan wrapper veya adapter başka package'a taşınarak yaşatılmaz.

Yeni Goose runner temiz ve küçük olarak yazılır. Eski runner dosyasını yeniden adlandırmak, içindeki eski policy'yi temizleyip aynı yapıyı korumak veya eski executor modelini Goose çağrısına bağlamak yasaktır.

Bu kural, Go source içinde eski migration sisteminin hiçbir parçasının kalmaması anlamına gelir. Domain/application kodu yalnızca migration altyapısından bağımsız olduğu ve halen ürün tarafından kullanıldığı kanıtlanırsa korunur.

## 5. SQL ve PostgreSQL kalıntıları

Yeni database'in şeması yalnız Foundation ve Goose tarafından oluşturulur. Eski migration framework'ünün SQL nesneleri yeni database'e dahil edilmez.

Repository'den silinecek SQL/script kalıntıları:

- Eski migration metadata tablolarını oluşturan SQL.
- Migration catalog, manifest, release, generation veya history tabloları.
- Migration run/attempt/checkpoint/receipt/requirement tabloları.
- Eski Goose completion routine/procedure/function/trigger'ları.
- Migration selector veya metadata için kullanılan view'lar.
- Eski migrator'a ait grant/ownership bootstrap SQL'i.
- Eski migration SQL corpus'u ve superseded/duplicate migration dosyaları.
- Eski schema'yı baseline/adoption için kullanan SQL.
- Eski framework'ün oluşturduğu state'i okuyan veya güncelleyen script'ler.

Bunların yerine eski metadata'yı yeni Goose migration'ıyla temizleyen bir işlem yazılmaz. Çünkü yeni database fresh başlar ve eski metadata bu database'de hiç oluşturulmaz. Eski database yeni kurulumun girdisi değildir.

SQL aramasında eski nesne adı başka bir domain tablosuyla çakışıyorsa bağlamı doğrula. Domain nesnesi olduğu kanıtlanan bir tabloyu yalnız isim benzerliği nedeniyle silme; ancak migration framework state'i olduğu doğrulanan nesne yeni schema tasarımına taşınmaz.

## 6. Contract ve generated output temizliği

Aşağıdaki tüm referanslar kaldırılır:

- Migration-specific Go contract'ları.
- Migration manifest/release JSON Schema dosyaları.
- Migration run/receipt/checkpoint payload'ları.
- Binary schema contract ve migration readiness contract'ları.
- Migration job/claim/lease API payload'ları.
- Bu contract'ları kullanan REST/RPC/message schema'ları.
- Generator source, generator config, generated output ve snapshot'lar.
- Silinen contract'ları test eden validation kodları.

Yalnız generated output'u silmek yeterli değildir. Generator tekrar çalıştırıldığında bu dosyaları üretmemelidir. Consumer, schema, test, API response, documentation ve deployment referansları birlikte temizlenir.

Yeni Goose runner için migration manifest veya özel contract oluşturulmaz. Runner'ın ihtiyaç duyduğu bağlantı ve source parametreleri normal Go config/CLI sınırında tanımlanır; migration state taşıyan bir contract oluşturulmaz.

## 7. CLI, Taskfile, container ve deployment temizliği

Repository genelinde eski migration sistemini çalıştıran veya paketleyen bütün wiring kaldırılır.

### CLI ve executable

- Eski migrator binary/command silinir.
- Migration-specific subcommand ve flag'ler kaldırılır.
- Eski platform/tenant migrator entrypoint'leri kaldırılır.
- Eski status/verify/reconcile/finalize/catalog/release komutları kaldırılır.
- Eski manifest/archive input/output parametreleri kaldırılır.

Yeni kurulum için yalnız tek INITIALIZE entrypoint'i kalır. Goose'u doğrudan geliştiriciye ikinci bir production bootstrap yolu olarak sunma; lifecycle çağrısı içinde kullan.

### Docker ve Compose

- Eski migrator image/build target kaldırılır.
- Manifest/archive dosyası COPY adımları kaldırılır.
- Migrator container/service kaldırılır.
- Tenant executor container/service kaldırılır.
- Eski healthcheck, depends_on, volume, port ve network referansları kaldırılır.
- Eski environment variable ve secret tanımları kaldırılır.
- Eski binary veya artifact runtime image içine kopyalanmaz.

### Taskfile ve script

- Eski migration build/copy/verify/catalog/release task'ları kaldırılır.
- Duplicate initialize/bootstrap/migrate task'ları kaldırılır.
- Eski SQL corpus/manifest üreten script'ler kaldırılır.
- Local development ve CI yalnız yeni INITIALIZE akışını kullanır.
- Yeni Goose source'u dışında ikinci bir SQL uygulama script'i bırakılmaz.

### CI ve release

- Eski migration manifest validation job'ları kaldırılır.
- Eski release/digest/generation doğrulama job'ları kaldırılır.
- Tenant executor deployability kontrolü kaldırılır.
- Eski binary/artifact üretimi kaldırılır.
- Yeni dört aşamalı akışın testleri gerçek PostgreSQL üzerinde çalıştırılır.

## 8. Dokümantasyon ve skill temizliği

Repository'de aktif talimat niteliği taşıyan bütün Markdown dosyaları kontrol edilir:

- AGENTS.md ve alt klasörlerdeki AGENTS.md dosyaları.
- skills/INDEX.md.
- skills/data-tenancy/SKILL.md.
- skills/platform-runtime/SKILL.md.
- Test ve impact skill'leri.
- docs/architecture altındaki database, platform, tenancy, runtime ve deployment belgeleri.
- README ve local development guide.
- CI/deployment/operations rehberleri.
- Eski migration task, archive ve plan klasörleri.

Eski sistemi öneren veya tanımlayan normatif metinler kaldırılır ve güncel dört aşamalı tasarımla yeniden yazılır. Eski migration task/archive klasörleri aktif dokümantasyon olarak tutulmaz.

Bu task setindeki yasaklı kavram listeleri yalnız silme kontrol listesi görevi görür. Bunlar yeni sistemin parçası, uygulanacak model veya korunacak API olarak yorumlanamaz.

## 9. Repository çapında tarama

Root dizininden case-sensitive ve case-insensitive arama yap. Sadece kelime aramasıyla yetinme; package/import graph, symbol reference, config key, SQL object ve deployable ilişkilerini de takip et.

Arama aileleri:

- migration, migrator, migrator executor, migration worker
- manifest, bundle, archive, catalog, selected migration
- digest, hash, release, generation, attestation
- run, attempt, checkpoint, receipt, requirement
- baseline, adoption, version override, custom history
- alias, supersession, classification, selector, planner
- schema contract, binary contract, readiness, finalizer
- Goose custom store/version store
- tenant migration service/executor/claim/lease
- eski migration routine/table/function isimleri ve eski CLI flag'leri

Arama sonuçları dosya bazında incelenir. Her eşleşme şu üç durumdan biriyle kapatılır:

1. Eski migration framework kalıntısıdır: silinir.
2. Yeni Goose implementation'ının zorunlu ve minimal bir parçasıdır: kapsamı doğrulanır ve tutulur.
3. Migration dışı, gerçek application/domain davranışıdır: neden tutulduğu belgelenir.

Belirsiz eşleşme otomatik olarak tutulmaz. Consumer ve işlevi araştırılır; eski framework'e ait olduğu ortaya çıkarsa tamamen kaldırılır.

## 10. Güvenli fresh database doğrulaması

Testler için eski database kullanılmaz. Ayrı, boş ve disposable PostgreSQL database/container oluşturulur.

- Test başlangıcında database gerçekten boş olmalıdır.
- Eski migration tablosu veya application schema'sı fixture olarak yüklenmez.
- Eski SQL corpus'u test setup içinde çalıştırılmaz.
- Goose history elle seed edilmez.
- Baseline/adoption simülasyonu yapılmaz.
- Test yalnız INITIALIZE → FOUNDATION → GOOSE → APPLICATION sırasını izler.

Test ortamını resetlemek gerekiyorsa yalnız test için ayrılmış disposable database/container yeniden oluşturulur. Geliştirici veya production database'i otomatik silen genel bir reset komutu eklenmez.

## 11. Zorunlu uçtan uca kabul testleri

### A. Fresh install

1. Tamamen boş PostgreSQL database/container hazırla.
2. INITIALIZE çalıştır.
3. DATABASE FOUNDATION yalnız altyapıyı oluştursun.
4. Goose platform ve tenant başlangıç schema'larını oluştursun.
5. Goose native history tabloları Goose tarafından oluşsun.
6. APPLICATION başlasın.
7. API/worker readiness başarılı olsun.

### B. Tekrar çalıştırma

1. Aynı yeni kurulumda INITIALIZE tekrar çağrılsın.
2. Foundation duplicate nesne oluşturmasın.
3. Goose tamamlanmış migration'ları tekrar çalıştırmasın.
4. Domain tablolarda duplicate seed/data oluşmasın.
5. Hiçbir eski state tablosu veya custom checkpoint kullanılmasın.

### C. Tenant davranışı

1. İki ayrı tenant için boş schema oluştur.
2. Aynı Goose runner ile ikisini de migrate et.
3. Her tenant'ın Goose history'si kendi schema kapsamına yazılsın.
4. Tenant A işlemleri Tenant B schema'sını değiştirmesin.
5. Migration başarısız tenant aktif edilmesin.
6. Tenant'ın tekrar kurulumu için custom migration state gerekmesin.

### D. Yetki ve startup

1. API/worker role ile CREATE TABLE denemesi başarısız olsun.
2. API/worker role schema/extension/role yönetemesin.
3. Goose role gerekli DDL işlemlerini gerçekleştirsin.
4. Goose başarısız olduğunda API/worker başlatılmasın.
5. API/worker restart'ı migration çalıştırmasın.

### E. Eski sistem yokluğu

1. Yeni database'de eski migration framework nesnesi oluşmadığını doğrula.
2. Go source ve Go test'lerde eski migration symbol/package/import kalmadığını doğrula.
3. Docker/Compose runtime'ında eski migrator deployable bulunmadığını doğrula.
4. Taskfile/CI içinde eski migration komutu bulunmadığını doğrula.
5. Contract/generated output içinde eski migration payload'ı bulunmadığını doğrula.
6. Eski migration SQL corpus'unun yeni source'a taşınmadığını doğrula.
7. Aktif architecture/skill docs'unun eski modeli önermediğini doğrula.

## 12. Çalıştırılacak kontroller

- **go test ./...**
- **go vet ./...** (repository standardında kullanılıyorsa)
- **go run ./tools/fileguard**
- **go run ./tools/archguard**
- Contract ve generated output doğrulamaları
- Goose platform integration testleri
- Goose tenant integration testleri
- Foundation permission integration testleri
- Fresh initialize integration testleri
- **task dev** ile boş PostgreSQL üzerinde uçtan uca kurulum
- Repository genelinde eski symbol/package/config/SQL/deployable taraması

Kontrollerden biri çalıştırılamazsa başarılı raporlanmaz. Çalıştırılamama nedeni ve kalan blocker açıkça yazılır. Guard veya test, yeni mimariye uyması için gevşetilmez; gerekiyorsa yalnız artık geçersiz legacy kuralı kaldırılır ve yeni dört aşamalı kural eklenir.

## 13. Nihai teslim raporu

Tamamlanma raporunda şu kanıtlar bulunmalıdır:

- Dört aşamanın gerçek entrypoint ve çağrı sırası.
- Sıfırdan yazılmış INITIALIZE/Foundation/Goose bağlantısı.
- Yeni Goose migration source dizinleri.
- Platform ve tenant native Goose history test sonucu.
- Yeni database'in ilk kurulum öncesi boş olduğunun kanıtı.
- API/worker DDL yetkisinin bulunmadığının kanıtı.
- Eski Go package/symbol/import temizliğinin sonucu.
- Eski SQL/contracts/generated output temizliğinin sonucu.
- Eski CLI/Taskfile/Compose/Docker/CI wiring temizliğinin sonucu.
- Eski docs/skills/archive temizliğinin sonucu.
- Çalıştırılan tüm test ve guard komutları.
- Başarısız veya atlanan testlerin açık listesi.

## 14. Kesin tamamlanma kriteri

İş aşağıdaki maddelerin tamamı sağlandığında biter:

- Runtime lifecycle yalnız dört aşamadır.
- Yeni sistem boş database üzerinde sıfırdan çalışır.
- Eski database/schema/migration state yeni sisteme alınmaz.
- Goose tek schema migration otoritesidir.
- Goose native history dışında migration state bulunmaz.
- Eski migration sisteminden hiçbir executable Go kodu kalmaz.
- Eski SQL corpus'u, contract, generated output, config, CLI, build artifact ve deployable kalmaz.
- Eski framework'ü tanıyan/adopte eden/temizleyen compatibility veya transition logic kalmaz.
- API/worker migration yapmaz ve DDL yetkisi taşımaz.
- Testler ve guard'lar gerçek sonuçlarıyla raporlanmıştır.

**Eski sistemden en ufak executable kalıntı, eski migration dosyasından taşınmış yeni bir migration veya eski state'i kullanan tek bir compatibility yolu varsa görev BAŞARISIZDIR.**


## 15. Dosya ve consumer bazında zorunlu disposition envanteri

Bu task'ın uygulanmasında aşağıdaki package kümeleri topluca incelenir. Klasör silmek tek başına yeterli değildir; package dışı consumer, generated producer, config, SQL, test ve deploy wiring aynı değişiklik dizisinde kapatılır. Başlangıç sayıları 2026-11-03 main tree denetiminden gelir; uygulama anında güncel recursive tree ile tekrar sayılır.

| Path | Başlangıç sayısı / not | Beklenen sonuç |
| --- | ---: | --- |
| `core/internal/platform/migrator/` | 289 dosya | Goose çağrısı dışındaki custom framework tamamen silinir. Backup/restore dışarı ayrılmadan package silinmez. |
| `core/internal/composition/migrator/` | 48 dosya | Migrator composition silinir; backup/restore kendi composition/command'ına taşınır. |
| `core/internal/platform/migration/` | 39 dosya | Migration document/release/catalog model'leri silinir. |
| `core/internal/platform/baseline/` | 24 dosya | Snapshot provider ve custom baseline lifecycle silinir. |
| `core/internal/platform/tenantmigrationexecutor/` | 8 dosya | Tenant executor process/state silinir. |
| `core/internal/platform/schemacontract/` | 16 dosya | Migration schema contract kaldırılır; runtime DB reachability/role check'i DDL'siz yeniden kurulur. |
| `core/internal/platform/migrator/tenantmigrationexecutor/` | 16 dosya | Alt executor adapter/SQL/Go kodu silinir. |
| `core/internal/platform/migratoroutput/` | repository tree'de mevcut | Plan/up/verify output payload'ları silinir. |
| `tools/localrelease/` | 66 dosya | Migration release/manifest işlevleri silinir; installation/license/release dışı işler ayrı owner'a taşınır. |
| `tools/internal/migrationstreamcatalog/` | 19 dosya | Catalog reader/generator/test'leri silinir. |
| `internal/attestation/migrationsupersession/` | 14 dosya | Supersession attestation modeli ve fixture'ları silinir. |
| `core/internal/modules/tenancy/adapters/postgres/tenantmigration/` | 11 dosya | Migration adapter kaldırılır; tenant domain provisioning korunur. |
| `core/cmd/migrator/` | entrypoint | Binary kaldırılır. |
| `core/internal/platform/migrator/schemaexecutor/` | repository tree'de mevcut | Typed DDL IR executor tamamen kaldırılır; dynamic form DDL işi Goose-only modeline göre yeniden tasarlanır. |

Her package için PR'da `path`, dosya sayısı, silinen dosya listesi, package dışı consumer listesi, yeni owner/path (varsa), test kanıtı ve kalan eşleşme disposition'ı bulunur. Başlangıç sayıları nihai sayım değildir.

## 16. Backup/restore'un migrator'dan bağımsızlaştırılması

`core-migrator` bugün yalnız schema migration yapmıyor. Aynı CLI/composition içinde şu migration dışı işler de bulunuyor:

- `database-backup-executor`
- `database-restore`
- `cutover_adoption` ve `cutover_finalization` (bunlar migration framework/cutover bağımlılıklarıdır; yeni modele taşınmaz, ilgili domain operasyonları ayrı değerlendirilir)

İlk iki işlev migration engine değildir ve silinmeleri recovery'yi bozar. Bunlar:
1. Goose veya migration package import etmeyen ayrı Go package/command/composition altında yeniden sahiplenilir.
2. Backup manifest, recovery set, signing/encryption, restore-plan validation, quiesce/drain, restore readiness, reactivation ve license reconciliation behavior'ını korur.
3. `core-migrator` binary'sine alias/subcommand olarak geri eklenmez.
4. Ayrı binary adı, config profile, least-privilege PostgreSQL role ve secret setine sahip olur.
5. Compose, Dockerfile, `tools/localconfig` artifact producer/catalog, secret projector, volume inventory, Taskfile, healthcheck, logs/status, recovery tests ve documentation birlikte değiştirilir.
6. Restore sonrasında Goose otomatik çalıştırmaz ve restore edilen schema'yı adopt etmez.

Zorunlu testler: backup al/verify; aynı DB'ye zarar vermeden restore planı reject; disposable hedefe restore; backup/restore executable'ının Goose package'larına bağımlı olmadığının archguard testi; `rehydrate` akışının boş hedef preflight ve recovery ordering testi.

## 17. Taskfile exact graph dönüşümü

Güncel ana akış `.taskfiles/local.yml` içindeki `local:up` görevidir. Şu sıralama tek bir initialize çağrısına dönüştürülür:

`prepare → infra → bootstrap → migrate → installation:artifacts → license → tenant-executor → initialize → services → verify → frontend:if-present → status`

Yeni Taskfile'da:
- `task dev` mevcut root alias üzerinden tek `local:up` görevini çağırmaya devam edebilir.
- `prepare` ve `infra` migration dışı hazırlık olarak kalabilir.
- `bootstrap` ile `migrate` ayrı task olmaktan çıkar; legacy `local-release prepare`, `migration-manifest`, `core-migrator deploy up`, `grant-runtime`, `core-migrator deploy verify` zinciri kaldırılır.
- `tenant-executor` ve schema executor task/service/healthcheck'i kaldırılır.
- Tek `initialize` görevi Foundation + Goose sırasını çağıran gerçek yeni entrypoint'i çalıştırır.
- Installation artifact/license/delivery adımları ancak yeni initialize dependency contract'ı ile çelişmiyorsa korunur; migration release state'ine bağlanamaz.
- `services` yalnız `core-api`, `core-worker` ve migration dışı servisleri başlatır.
- `verify` yalnız gerçek runtime readiness kontrol eder.
- `frontend:if-present` ve `status` davranışı korunur.

`.taskfiles/local-recovery.yml` içindeki `rehydrate` görevi şu an `:local:bootstrap` ve `:local:migrate` çağırır. Bu iki adım tek initialize girişine çevrilir. Ancak restore kit'in DB içeriği restore edilmeden önce Goose çalıştırmak ile önce restore edip sonra initialize yapmak farklı sonuç doğurabilir; task uygulayıcısı tahminle sıra seçemez. İki yaklaşımı disposable PostgreSQL üzerinde test eder, recovery set formatı ve restore semantics ile uyumlu olanı seçer ve ordering testini ekler. Eski `bootstrap/migrate` task'larını alias bırakmak yasaktır.

`.taskfiles/local-operations.yml` log/status/check görevleri `core-migrator`, `schema-executor`, `tenant-migration-executor` adlarını çıkarır. Backup/restore executable adları gerçek yeni binary isimleriyle gösterilir.

`ops/local/taskfile_contract_test.go` ve `ops/local/recovery_contract_test.go` yeni task graph'i doğrular. Eski `core-migrator deploy verify` beklentileri ve executor readiness URL'leri kaldırılır; testler yalnız string absence değil task dependency/order davranışını da doğrular.

## 18. Container ve artifact graph kesin temizliği

Aşağıdaki güncel dosyaların tüm producer/consumer referansları kapatılır:

- `ops/local/Dockerfile`
- `ops/production/Dockerfile`
- `ops/local/compose.tools.yml`
- `ops/local/compose.runtime.yml`
- `ops/local/compose.recovery.yml`
- `ops/local/compose.lifecycle.yml`
- `ops/local/compose.inventory.yml`
- `.taskfiles/local.yml`
- `.taskfiles/local-recovery.yml`
- `.taskfiles/local-operations.yml`
- `tools/localconfig/profile_environment_artifacts.go`
- `tools/localconfig/operation_environment.go`
- `tools/localconfig/artifact_catalog.go`
- `tools/localconfig/profile_environment_test.go`
- `cmd/setup/release_material_keep.go`

Eski `migrator.env`, `schema-executor.env`, `tenant-migration-executor.env`, `HARPY_MIGRATION_MANIFEST_PATH`, migration manifest file, release-state volume, migrator/schema secret volumes, executor ports, health URL'leri ve eski image binary'leri artifact producer'dan başlayarak kaldırılır. Yalnız Compose'tan service satırı silmek yeterli değildir; generated artifact set exact list ve secret inventory testleri stale dosya üretmediğini kanıtlar.

## 19. Contract ve SQL nesne disposition zorunluluğu

Aşağıdaki dosyalar migration framework contract'ı olarak doğrudan incelenir ve consumer/generator ilişkisi kapatılarak kaldırılır:

- `contracts/postgres/v1/migration-manifest.schema.json`
- `contracts/postgres/v1/migration_source.go`
- `contracts/postgres/v1/migration_types.go`
- `contracts/postgres/v1/select_migration.go`
- `contracts/postgres/v1/validate_migration.go`
- `contracts/runtime/v1/core-migrator-config.schema.json`
- `contracts/runtime/v1/fresh-migration-manifest.json`
- `contracts/runtime/v1/fresh_migration_manifest.go`
- `contracts/runtime/v1/binary-schema-contract.json`
- `contracts/runtime/v1/binary_schema_contract.go`
- `contracts/runtime/v1/migrator-command-result.schema.json`
- `contracts/runtime/v1/migrator_command_result.go`
- `contracts/storage/migration-streams.yaml` ve include ettiği JSON dosyaları
- `contracts/storage/owners/platform_migrator.json`

Eski platform migrator owner envanterinde listelenen 17 platform table, 39 routine ve external access/grant kayıtları yeni DB tasarımı için otomatik truth kabul edilmez. Her biri SQL object inventory ve güncel consumer üzerinden disposition alır. Migration run/checkpoint/release/catalog/routine nesneleri kaldırılır. `environment_identity` gibi application/security amacı da olabilecek nesneler isim benzerliğiyle silinmez; ayrı owner altında yeniden gerekçelendirilir.

Yeni Goose schema kapsamı, 14 migration-stream owner kaynağının her birinde tarihsel stream kayıtlarını güncel domain/storage/API consumer'larıyla çapraz kontrol ederek tamamlanır. Yeni schema nesneleri için owner, yeni Goose migration path'i ve gerçek PostgreSQL test adı zorunludur. Eski migration SQL corpus'u, stream order veya baseline dump yeni migration dosyasına kopyalanamaz.

## 20. Dynamic form DDL: ürün davranışı kapanışı

`schema-executor` ve typed DDL IR kaldırılması yalnız runtime service silmek anlamına gelmez. Bugünkü form/config publish ile fiziksel tenant schema nesnesi üretme davranışının ürün karşılığı açıkça kapatılmalıdır.

- Runtime publish sırasında Goose source oluşturulamaz/değiştirilemez/çalıştırılamaz.
- Publish path'inde DDL, CREATE/ALTER TABLE veya dynamic schema execution bulunamaz.
- Form metadata, field validation, record CRUD, search/filter/indexing ve tenant isolation için yeni storage design hazırlanır.
- JSONB/EAV veya tek tablo generic model otomatik seçilemez; gerçek API/query gereksinimleri ve performans/constraint trade-off'ları dokümante edilir.
- Eski typed DDL contract/IR payload'ı, schema-executor queue/command, secrets, config ve tests kaldırılır.
- Form publish, concurrent publish, record write/read, validation, query filter, permission ve isolation PostgreSQL integration testleri olmadan bu başlık tamamlanmaz.

## 21. CI, guard ve fresh install acceptance gate

`.github/workflows/ci.yml` bugünkü `baseline-pg18` job'unda Catalog cleanup snapshot testlerini çalıştırıyor. Bu job yeni PostgreSQL 18 fresh-install doğrulamasına dönüştürülür. Aşağıdaki CI kapıları zorunludur:

1. Ephemeral PostgreSQL 18 + pgvector boş DB kurulum testi.
2. Aynı DB üzerinde INITIALIZE idempotency testi.
3. Tenant A/B aynı Goose runner ve native history isolation testi.
4. API/worker role DDL negative permission testi.
5. Foundation failure ve Goose failure sonrası Application'ın başlamadığını doğrulayan orchestration testleri.
6. Goose uygulanmış migration dosyaları için Git diff immutability gate.
7. Bilinen legacy package/path/env/command/service/secret token'ları için source/deploy guard. Domain kelimelerini körlemesine yasaklama; her istisna gerekçe ve owner taşır.
8. Backup/restore command'ın migrator/Goose import graph dışında olduğunu doğrulayan archguard.
9. `go test ./...`, `go vet ./...`, integration compile, `go run ./tools/fileguard`, `go run ./tools/archguard`.
10. Disposable Compose üzerinde yeni initialize entrypoint'iyle gerçek local startup smoke; Docker runner yoksa bu kapı skipped/success sayılamaz.

Yeni guard kendisini de test eder: legacy fixture eklendiğinde fail, açıklamalı non-migration domain fixture eklendiğinde pass, Goose native source eklendiğinde pass, uygulanmış migration değiştirilince fail, yeni migration eklenince pass.

## 22. Sıfır kalıntı disposition kayıt biçimi

Her eski eşleşme repository içi bir audit artifact'inde aşağıdaki alanlarla kaydedilir:

| Alan | Açıklama |
| --- | --- |
| `path` | Tam repository yolu |
| `symbol_or_object` | Go symbol, SQL object, env key, service, task veya contract adı |
| `legacy_owner` | Eski package/owner |
| `consumers` | Import/call/config/generated/deploy/docs consumer'ları |
| `decision` | `REMOVED`, `REPLACED_BY_GOOSE`, `PRESERVED_NON_MIGRATION`, `BLOCKED` |
| `replacement` | Yeni path/symbol veya boş |
| `test_evidence` | Test/guard adı ve sonucu |
| `reason` | Koruma/silme gerekçesi |

`BLOCKED`, boş consumer, testsiz replacement veya gerekçesiz preserve satırı varken sıfır kalıntı tamamlanmış değildir. Audit artifact'i eski migration system'in runtime contract'ı değildir; yalnız PR denetim kanıtıdır ve migration manifest/catalog formatına dönüştürülmez.

## 23. Son kullanıcı güvenliği / mevcut volume kuralı

`task dev`, initialize veya recovery akışı mevcut PostgreSQL named volume'u otomatik silmez, yeniden oluşturmaz, truncate etmez veya eski schema'yı temizlemez. `task local:destroy` yalnız açıkça kullanıcı tarafından çağrılan ve onay verilen local destructive operasyon olarak kalır. Yeni fresh install için ayrı disposable volume/database kullanılır. Eski geliştirme volume'u otomatik migrate/adopt edilmez; açık hata ve manuel ayrı boş database gereksinimi ile durulur.

