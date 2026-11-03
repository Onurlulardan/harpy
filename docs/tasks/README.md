# Harpy — Sıfırdan Veritabanı Yaşam Döngüsü

## 1. Bu task setinin değiştirilemez kararı

Harpy'nin veritabanı kurulum ve schema yönetimi mevcut migration altyapısının devamı değildir. Mevcut altyapı referans alınarak iyileştirilmeyecek, sadeleştirilmeyecek, Goose'a uyarlanmayacak ve bir uyumluluk katmanı altında yaşatılmayacaktır.

**Mevcut migration sistemi tamamen silinecek. Yeni sistem boş bir zeminde sıfırdan yazılacaktır.**

Bu karar yalnızca package isimleri için değil, davranış, veri modeli, SQL, sözleşme, konfigürasyon, build girdisi, deployable, test, fixture, CLI ve dokümantasyon için de geçerlidir. Eski sistemin bir parçasını yeni isimle taşımak, aynı davranışı başka bir struct veya servis arkasına saklamak ya da eski state'i yeni sisteme okutmak tamamlanmış iş sayılmaz.

Yeni sistemin tek yaşam döngüsü aşağıdaki dört aşamadır:

1. **INITIALIZE**
2. **DATABASE FOUNDATION**
3. **GOOSE**
4. **APPLICATION**

Bu dört aşamanın dışında beşinci bir lifecycle aşaması, paralel schema yöneticisi veya ayrı migration servisi bulunmayacaktır.

## 2. Fresh başlangıç kuralı

Yeni sistem yalnızca **boş bir PostgreSQL database'i** hedefler. Eski migration sisteminin oluşturduğu database, tabloları, schema'ları, metadata'sı, migration dosyaları veya version kayıtları yeni sisteme girdi olarak verilmeyecektir.

Aşağıdaki işlemler kesinlikle yapılmayacaktır:

- Eski database'i Goose geçmişine dahil etmek.
- Eski schema'yı yeni ilk migration ile eşleştirmek veya baseline/adoption yapmak.
- Eski migration geçmişini Goose version tablosuna yazmak.
- Eski SQL migration dosyalarını yeni migration source'una kopyalamak.
- Eski metadata'yı yeni migration çalıştırarak temizlemek.
- Eski veriyi, schema'yı veya migration state'ini yeni sisteme taşımak.
- Eski davranışla uyum sağlamak için adapter, fallback, translator veya compatibility mode oluşturmak.

Yeni kurulum için yeni ve boş bir database oluşturulur. Eski database yeni sistem tarafından okunmaz, değiştirilmez veya otomatik olarak silinmez. Konfigürasyon yanlışlıkla dolu/eski bir database'i gösteriyorsa kurulum **fail closed** davranışıyla durur. Eski database'i otomatik silmek de kabul edilmez; yeni kurulum açıkça ayrılmış boş bir database üzerinde yapılır.

Eski uygulama verisinin taşınması bu task setinin kapsamı değildir. Yeni database boş başlar. Bir veri aktarımı ihtiyacı doğarsa bu, migration altyapısına eklenmez ve bu task setinin tamamlanma koşulu yapılamaz.

## 3. Nihai mimari

### INITIALIZE

Tek kurulum giriş noktasıdır. Aşamaları doğru sırada çağırır, hata durumunda durur ve tekrar çalıştırılabilir. Kendi version, run, checkpoint veya completion state'ini tutmaz.

### DATABASE FOUNDATION

Boş PostgreSQL üzerinde database seviyesindeki zorunlu altyapıyı kurar: roller, schema namespace'leri, extension'lar, ownership, grants ve default privileges. Uygulama tabloları veya migration geçmişi oluşturmaz.

### GOOSE

Bütün platform ve tenant schema DDL'inin tek otoritesidir. Resmi **github.com/pressly/goose/v3** kullanılır. İlk schema, eski migration zincirinden devralınmaz; mevcut ürün gereksinimleri esas alınarak sıfırdan hazırlanmış yeni Goose başlangıç migration'larıyla oluşturulur. Sonraki bütün schema değişiklikleri yine Goose migration dosyalarıyla yapılır.

### APPLICATION

API, worker ve diğer uygulama süreçleridir. Foundation ve Goose başarıyla bitmeden başlatılmaz. Normal uygulama runtime'ı DDL veya database administration yapmaz.

## 4. Mimari sınırlar

- Runtime lifecycle tam olarak dört aşamadır.
- Tek migration otoritesi Goose'tur.
- Goose native version history dışında migration history yoktur.
- Platform ve tenant migration'ları aynı Goose implementation/runner ile yürütülür.
- Tenant schema için ikinci bir migrator, executor, worker, daemon veya migration-specific queue bulunmaz.
- Foundation domain tablolarını oluşturmaz.
- Application startup migration çalıştırmaz.
- Hiçbir eski migration Go kodu yeni runner'a taşınmaz.
- Eski migration SQL corpus'u yeni başlangıç migration'ı olarak kullanılmaz.
- Eski migration framework metadata'sı yeni database'e oluşturulmaz.
- Goose native version tablosu Foundation tarafından yaratılmaz veya yönetilmez.

Goose runner yalnızca hedef database bağlantısını, hedef schema/search_path değerini ve ilgili native migration source'unu Goose'a ileten küçük bir çağrı sınırıdır. Runner kendi discovery, sorting, selection, classification, versioning, retry state veya completion policy'sini geliştiremez.

## 5. İş sırası

1. [01 — INITIALIZE](01-initialize.md): eski entrypoint/orchestration kodunu kaldır ve tek yeni giriş noktasını sıfırdan oluştur.
2. [02 — DATABASE FOUNDATION](02-database-foundation.md): yalnız boş database için DB altyapısını oluştur.
3. [03 — GOOSE](03-goose.md): yeni platform ve tenant schema'larını temiz Goose migration source'larından sıfırdan üret.
4. [04 — APPLICATION ve sıfır kalıntı](04-application-and-total-removal.md): uygulama başlangıç sınırlarını kur, eski sistemi repository'nin tamamından sil ve kanıtla.

Task dosyaları iş sırasını belirtir. Bunlar runtime'a yeni aşama eklemez.

## 6. Repository çapındaki silme kapsamı

Arama ve kaldırma kapsamı repository root'un tamamıdır:

- Bütün Go source ve test dosyaları
- SQL dosyaları ve PostgreSQL routine'leri
- JSON/JSON Schema, YAML, protobuf ve diğer contract dosyaları
- Generated source ve generator'lar
- Test fixture, snapshot ve mock'lar
- Dockerfile, Compose, Taskfile ve shell script'leri
- CLI komutları ve flag'leri
- CI workflow ve release/build adımları
- Environment variable, config alanı, secret ve dosya path'leri
- Runtime image, binary, volume, service ve deployable tanımları
- AGENTS.md, skills, architecture docs, developer guide ve README'ler
- Eski task, archive ve migration planı klasörleri

Yalnızca bir dosyayı silmek yeterli değildir. O dosyanın import eden, üreten, çağıran, test eden, deploy eden veya dokümante eden bütün bağlantıları da kaldırılmalıdır. Generated output'lar yeniden üretilerek temizlenmelidir; elle silinip generator tarafından geri getirilmeleri kabul edilmez.

## 7. Yeni başlangıç schema'sının kaynağı

Yeni schema, eski migration dosyalarının kopyası değildir. Yeni başlangıç schema'sı mevcut ürünün bugün desteklenmesi gereken domain ve platform gereksinimleri, güncel application contract'ları ve açıkça onaylanmış veri modeli üzerinden yeniden tasarlanır.

- Eski migration dosyaları schema tasarımının kaynağı yapılmaz.
- Eski SQL corpus'undaki dosyalar yeni Goose dosyalarına taşınmaz.
- Eski migration sıralaması veya eski version numaraları korunmaz.
- Eski migration metadata tabloları yeniden yaratılmaz.
- Gerekli tablo, index, constraint, function, trigger ve view'lar yeni SQL olarak yazılır.
- Kullanılmayan, çakışan veya tarihsel olarak kalmış schema nesneleri yeni başlangıç schema'sına alınmaz.
- Uygulama kodu ile yeni schema uyuşmuyorsa hangi güncel domain davranışının destekleneceği belirlenir; eski migration davranışı otomatik olarak doğru kabul edilmez.

## 8. Test ve tamamlanma ölçütü

En az aşağıdaki doğrulamalar gerçek PostgreSQL üzerinde yapılır:

- Tamamen boş database ile ilk kurulum.
- INITIALIZE tekrar çalıştırıldığında idempotent sonuç.
- Foundation'ın yalnız DB altyapısını oluşturması.
- Goose'un platform schema'sını sıfırdan oluşturması.
- İki ayrı tenant schema'sının aynı Goose runner ile kurulması.
- Goose hata verdiğinde Application'ın başlamaması.
- API/worker rollerinin DDL yapamaması.
- API/worker restart'ında migration çalışmaması.
- Tenant izolasyonu.
- Tekrarlanan Goose çağrısının yeni DDL/data kopyası üretmemesi.
- Eski migration davranışı veya eski migration state'i olmadan clean install.
- Repository genelinde eski runtime/code/config/deployable kalıntısı bulunmadığının kanıtlanması.

Çalıştırılacak kontroller:

- **go test ./...**
- **go run ./tools/fileguard**
- **go run ./tools/archguard**
- Repository'de tanımlı contract ve generated output kontrolleri
- **task dev** ile tamamen boş database üzerinde uçtan uca kurulum
- Gerçek PostgreSQL tenant ve permission integration testleri

Docker veya PostgreSQL erişimi yoksa ilgili test başarılı varsayılmaz; açık blocker olarak raporlanır.

**Bir eski migration kodu, eski migration SQL'i, eski migration contract'ı, eski migration config'i, eski migrator deployable'ı veya eski migration davranışı kalmışsa task seti tamamlanmamıştır.**


## 9. Repository envanteri ve disposition

Bu bölüm soyut bir arama önerisi değildir. Aşağıdaki envanter, `main` dalında 2026-11-03 tarihinde doğrulanan mevcut dosya kümelerini ve bunların zorunlu disposition'ını tanımlar. Sayılar sabit kabul edilmez; uygulamaya başlamadan önce yeniden sayılır ve PR raporuna güncel sayılar yazılır.

### 9.1 Bilinen eski runtime kümeleri

| Yol | Doğrulanmış mevcut kapsam | Zorunlu disposition |
| --- | ---: | --- |
| `core/internal/platform/migrator/` | 289 dosya | Goose runner dışındaki tamamı kaldırılır. Backup/restore sorumlulukları ayrıştırılmadan dosyalar silinmez. |
| `core/internal/composition/migrator/` | 48 dosya | Tamamı kaldırılır; backup/restore composition'ı ayrı, migration dışı package'a taşınır. |
| `core/internal/platform/migration/` | 39 dosya | Custom migration document/release modelinin tamamı kaldırılır. |
| `core/internal/platform/baseline/` | 24 dosya | Özel baseline provider/snapshot orkestrasyonu kaldırılır; yalnız yeni native Goose source ve minimal çağrı sınırı kalır. |
| `core/internal/platform/tenantmigrationexecutor/` | 8 dosya | Tamamı kaldırılır; tenant provisioning aynı Goose runner ile synchronous çağrı yapar. |
| `core/internal/platform/schemacontract/` | 16 dosya | Migration/release readiness contract'ı kaldırılır; runtime schema readiness DDL/version yönetmeden gerekli uygulama bağlantılarını sınar. |
| `core/internal/platform/migrator/tenantmigrationexecutor/` | 16 dosya | Tamamı kaldırılır. |
| `tools/localrelease/` | 66 dosya | Migration manifest/release-ledger kısmı kaldırılır; installation/license gibi migration dışı işlevler gerçek owner'a taşınır ve test edilir. |
| `tools/internal/migrationstreamcatalog/` | 19 dosya | Tamamı kaldırılır. |
| `internal/attestation/migrationsupersession/` | 14 dosya | Tamamı kaldırılır. |
| `core/internal/modules/tenancy/adapters/postgres/tenantmigration/` | 11 dosya | Migration executor adapter'ı kaldırılır; tenancy domain provisioning yalnız domain state'ini yönetir. |
| `core/cmd/migrator/` | 1 entrypoint | Tamamı kaldırılır. |
| `core/internal/platform/migrator/schemaexecutor/` | repository ağacında mevcut | Tamamı kaldırılır; form/config DDL publish akışı da Goose-only karara göre yeniden tasarlanır. |
| `core/internal/platform/migratoroutput/` | repository ağacında mevcut | Custom plan/up/verify payload üretimi kaldırılır. |

Bu tablo silinecek dosyaların tamamı değildir. Her package altındaki bütün dosyalar, alt dizinler, testler, SQL ve generated dosyalar PR'da listelenir. Paket dışındaki consumer ve üreticiler de aynı PR kapsamındadır.

### 9.2 Güncel baseline zinciri için özel karar

Mevcut `core/internal/platform/baseline/apply.go`, resmi Goose Provider kullansa da şu an dört baseline ailesini ardışık olarak uygular: eski platform snapshot, eski tenant snapshot, Catalog-removal snapshot'ları ve 2026-11-03 tenant-executor completion/claim-clock contract snapshot'ları (sonuncular `20261103000003`–`20261103000006`). Bu kod yeni hedefte korunmayacaktır.

- Sekiz mevcut `*.fresh.sql` snapshot dosyası yeni Goose source'una kopyalanmayacak.
- Eski version ID'leri yeni initial migration version'ı olarak yeniden kullanılmayacak.
- Snapshot'lar yalnız mevcut canlı ürün schema'sı için nesne/kolon/constraint/routine envanteri çıkarma girdisidir; SQL metinleri veya migration sırası taşınamaz.
- Yeni initial schema, güncel domain/application contract'larından bağımsız şekilde doğrulanır; her tablo, index, constraint, type, function, trigger, view ve policy için owner ve kullanım kanıtı tutulur.
- `20261103000003`–`20261103000006` ile getirilen tenant executor register/claim/heartbeat/complete/fail ve onboarding geçişi rutinleri yeni sisteme taşınmaz. Onboarding/tenant activation iş kuralı gerekiyorsa migration kavramı kullanmadan tenancy/application domain akışında yeniden tasarlanır.
- Halihazırda uygulanmış migration dosyaları geriye dönük değiştirilmez. Bu karar yeni ve boş database hedefi içindir; mevcut geliştirme volume'u bu task kapsamında otomatik resetlenmez.

### 9.3 Eski SQL stream envanterinin kullanımı

`contracts/storage/migration-streams.yaml` 14 owner dosyasını ve önceki denetimde sayılmış 312 stream kaydını birleştirir. Bu kayıtlar yeni migration source'u, sıralama planı, import listesi veya Goose history girdisi değildir.

Her owner dosyası yalnızca tarihsel schema kapsamı açısından incelenir. İçindeki her stream için:
1. Ürettiği PostgreSQL nesneleri belirlenir.
2. Nesnenin halen güncel application/domain contract'ı tarafından kullanılıp kullanılmadığı doğrulanır.
3. Kullanılıyorsa yeni başlangıç schema'sında yeni SQL olarak tasarlanır.
4. Kullanılmıyorsa yeni schema'ya alınmaz.
5. Karar; owner, nesne türü/adı, güncel consumer, yeni Goose migration dosyası veya kaldırma gerekçesiyle disposition tablosuna yazılır.

14 owner dosyasının her biri tek tek kapanmadan schema kapsamı tamamlandı işaretlenemez. Stream kayıtlarının sayısı değişmişse mevcut branch'ten yeniden hesaplanır; önceki 312 sayısı kontrol toplamıdır, otorite değildir.

### 9.4 Platform migrator database nesne envanteri

`contracts/storage/owners/platform_migrator.json` içindeki platform table/routine/grant/FK/external-access kayıtları eski framework'ün kendi metadata modelidir. Bu owner dosyası yeni storage owner contract'ı olarak korunmaz.

- `platform_tables` altındaki eski migration state tablolarının tamamı yeni schema'dan çıkarılır.
- `environment_identity` gibi bir isim tek başına migration state olduğu anlamına gelmez; application/security gereksinimi varsa ayrı owner altında yeniden tanımlanması ve nedeninin belgelenmesi zorunludur.
- `platform_routines` altındaki migrator validation, checkpoint, completion, claim, register, heartbeat, supersession ve release-finalization rutinleri kaldırılır.
- `external_accesses` içindeki tenancy/billing/async/cutover bağlantıları consumer bazında kapatılır. Domain davranışının devam etmesi gerekiyorsa migrator rutinini tutmak yerine ilgili owner modülünde migration bağımsız API/SQL tasarlanır.
- Eski owner dosyasını boşaltıp bırakmak kabul edilmez; generator ve tüketicileriyle birlikte kaldırılır.

## 10. Taskfile ve local lifecycle dönüşüm kontratı

Güncel `Taskfile.yml`, `.taskfiles/local.yml`, `.taskfiles/local-recovery.yml` ve `.taskfiles/local-operations.yml` birlikte okunur. Aşağıdaki eski akışın her task'ı için tek yeni karşılık veya açıkça kaldırma kararı verilmeden task seti tamamlanmış sayılamaz.

Mevcut `local:up` sırası:

```text
prepare → infra → bootstrap → migrate → installation:artifacts
→ license → tenant-executor → initialize → services → verify
→ frontend:if-present → status
```

Hedef davranış:
- `task dev` tek INITIALIZE girişini çağırır; Foundation ve Goose'u bu giriş çalıştırır.
- Ayrı `bootstrap`, `migrate`, `tenant-executor`, `schema-executor`, `migration:verify` veya `core-migrator deploy up/verify` adımı kalmaz.
- Installation artifact/license/delivery gibi migration dışı işlemler lifecycle içinde kalacaksa bunların migration state'e bağımlı olmayan gerçek owner ve sıraları açıkça tanımlanır.
- API/worker, Goose başarı sinyali geldikten sonra başlatılır.
- Tekrarlanan `task dev` migration harici installation/delivery işlemlerinin idempotency ve one-time delivery davranışını bozmaz.

| Mevcut task / akış | Zorunlu karar |
| --- | --- |
| `:local:bootstrap` | `:local:initialize` içindeki tek yeni entrypoint'e alınır; eski release/manifest/grant/verify adımları silinir. |
| `:local:migrate` | Tamamı kaldırılır; Goose ikinci kez Taskfile'dan çağrılmaz. |
| `:local:tenant-executor` | Servis ve task kaldırılır; tenant provisioning aynı Goose kod yoluna alınır. |
| `:local:initialize` | Tek lifecycle girişine dönüşür; installation delivery gibi migration dışı işlemler açık alt-adım olarak kalabilir. |
| `:local:services` | Schema executor başlatmaz; yalnız API, worker ve migration dışı runtime servislerini başlatır. |
| `:local:verify` | API/worker ve gerçek dependency readiness doğrular; migrator/executor endpoint'i beklemez. |
| `:local:up` / `task dev` | Yukarıdaki hedef sırayı uygular; tüm komutlar gerçek Taskfile task adlarıyla test edilir. |
| `:migration:verify` | Kaldırılır; Goose native sonucu ve integration tests dışında custom verify komutu yoktur. |
| `:local:recovery:rehydrate` | Eski bootstrap/migrate çağrılarını INITIALIZE girişine çevirir; önce boş hedef doğrulanır, restore ancak Foundation+Goose başarılıysa başlar. |
| `:local:recovery:restore` | Restore edilen veritabanını migration ile otomatik değiştirmez; restore sonrası kimlik/lisans reactivation işleri migration dışı recovery akışı olarak kalır. |
| `:local:operations:logs/status` | Migrator/schema/tenant executor isimlerini çıkarır; ayrı backup/restore process adlarını kullanır. |

`bootstrap` ile `migrate` bugün aynı `local-release prepare → migration manifest → core-migrator deploy up → grant-runtime → deploy verify` zincirini tekrar etmektedir. Yeni model bu tekrarın birini bırakıp diğerini alias yapamaz; zincirin tamamı kaldırılır.

## 11. Docker, Compose ve binary disposition

Aşağıdaki güncel dosyalar mutlaka düzenlenir; yalnız servis tanımını silmek yeterli değildir:

| Dosya | Şu anki legacy bağ | Hedef |
| --- | --- | --- |
| `ops/local/Dockerfile` | `core-migrator`, `local-release`, migration manifest build/COPY | Tek initialize + foundation/Goose implementation; migration dışı artifact işlevleri ayrı build target. |
| `ops/production/Dockerfile` | `core-migrator`, release-manager, migration-manifest üretimi | Production initialize/Goose sınırı; eski migrator artifact yok. |
| `ops/local/compose.tools.yml` | `core-migrator`, `local-release`, migrator secret/release volume | Tek initialize one-shot; installation/license işlevleri migration state kullanmadan bağımsız kalabilir. |
| `ops/local/compose.runtime.yml` | `tenant-migration-executor`, `schema-executor` service | İkisi de kaldırılır; API/worker yalnız initialize başarı sonrasında başlar. |
| `ops/local/compose.recovery.yml` | Backup/restore `core-migrator` subcommand'lerine bağlı | Backup/restore ayrı migration dışı executable ve secret/profile ile devam eder. |
| `ops/local/compose.lifecycle.yml` | Migrator, schema, release secret/state volume'ları | Sadece gerçekten kullanılan installation, app ve recovery volume'ları kalır. |
| `ops/local/compose.inventory.yml` | `/scan/release`, migrator DSN ve release state varsayımları | Yeni secret inventory kaynaklarıyla güncellenir; eski manifest/migrator secret taraması yok. |

Backup/restore şu anda `core-migrator` içinde `database-backup-executor` ve `database-restore` komutlarıdır. Bunları silip recovery'yi bozmak yasaktır. Bu işlevler:
- Migration import'u, Goose source'u, schema DDL'i veya migration history bağımlılığı taşımayan ayrı package/command olarak yeniden sahiplenilir.
- Backup/restore formatı, signing, encryption, restore-plan doğrulama, quiesce/drain, readiness ve reactivation contract'ları korunur.
- Backup/restore PostgreSQL data restore işlemi yapabilir; ancak schema migration veya eski migration state adoption işlemi yapamaz.
- Compose secret, volume, healthcheck, Taskfile, logs/status ve CI testleri yeni executable adına birlikte güncellenir.
- Eski migrator binary içinde ikinci entrypoint olarak bırakılması kabul edilmez.

## 12. Contract, config, test ve generator envanteri

Aşağıdaki dosyalar bilinen doğrudan kontrol noktalarıdır. Uygulama başlamadan önce güncel tree'den benzer dosyalar da çıkarılır.

### Contracts
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
- `contracts/storage/migration-streams.yaml`
- `contracts/storage/migration-streams/*.json`
- `contracts/storage/owners/platform_migrator.json`

Bunların tamamı consumer/generator/test ilişkileri çözüldükten sonra kaldırılır. Yalnızca tarihsel stream owner kaynakları schema disposition raporu üretilene kadar okunur; yeni migration source'u olmaz.

### Config ve artifact generator
- `tools/localconfig/profile_environment_artifacts.go`
- `tools/localconfig/profile_environment_test.go`
- `tools/localconfig/artifact_catalog.go`
- `tools/localconfig/operation_environment.go`
- `cmd/setup/release_material_keep.go`

Bu generator'lardan `migrator.env`, `schema-executor.env`, `tenant-migration-executor.env`, `HARPY_MIGRATION_MANIFEST_PATH`, release directory ve migration DSN/profile alanları kaldırılır. Generated artifact setleri tekrar üretilir; stale dosyalar yalnız generator çıktısı doğrulandıktan sonra silinir.

### Test ve CI
- `ops/local/taskfile_contract_test.go`
- `ops/local/recovery_contract_test.go`
- `.github/workflows/ci.yml`

Taskfile testleri yeni tek entrypoint ve dört lifecycle ordering'ini assert eder; eski `core-migrator deploy verify` beklentisini kaldırır. Recovery contract testleri yeni backup/restore executable'ını doğrular. CI PostgreSQL job'u eski Catalog cleanup snapshot testlerini değil, yeni Foundation/Goose/fresh-install testlerini koşar.

## 13. Dynamic form DDL ve ürün davranışı kararı

Bugünkü `schema-executor`, form/config publish sırasında typed DDL IR uygular. Bu davranış Goose-only kararıyla aynı anda korunamaz; çünkü runtime'da Goose dışı schema DDL yolu yaratır.

Bu task setinde tercih edilen hedef:
- Runtime form/config publish fiziksel tablo/kolon DDL'i uygulamaz.
- Form ve field tanımları uygulama metadata'sı olarak saklanır; record storage modeli mevcut ürün gereksinimleri temelinde DDL'siz çalışacak şekilde tasarlanır.
- Eğer mevcut domain modeli bu davranışı sağlayamıyorsa DDL'siz model ayrı tasarım/karar maddesi olarak çözülmeden schema-executor korunmaz.
- Dynamic DDL ihtiyacı yeni Goose migration gerektiriyorsa bu yalnız deployment/release-time kaynak migration değişikliği olarak yapılır; runtime publish migration dosyası üretip çalıştıramaz.
- Dynamic form alanı, validation, indexing ve query performansı gereksinimleri test edilmeden EAV/JSONB çözümü varsayılmaz. Nihai storage tasarımı ürün contract'ı ve gerçek consumer testleriyle belirlenir.

Bu ürün davranışını sessizce kaybetmek de kabul edilmez. Yeni modelde mevcut form publish, record CRUD, validation, search/filter, tenant isolation ve concurrent publish akışlarının karşılığı için test gerekir.

## 14. Catalog ve baseline snapshot geçmişi

Eski stream/owner katalogları ve Catalog-removal baseline'ları yeni sisteme taşınmaz. Yeni fresh schema içinde migration framework'üne ait tablo/routine bulunmadığı gerçek PostgreSQL catalog sorgularıyla doğrulanır.

Mevcut sekiz baseline snapshot'ın her biri yalnız envanter kaynağı olarak açılır. Her nesne için `KEEP_AS_PRODUCT_SCHEMA`, `RECREATE_AS_NEW_GOOSE_DDL`, `REMOVE_LEGACY_STATE` veya `UNRESOLVED_BLOCKER` disposition'ı yazılır. `UNRESOLVED_BLOCKER` varken initial schema tamamlanmış sayılmaz.

Catalog-removal çalışmasının sonucunu tekrar üretmek için yeni migration yazılmaz. Yeni database'de bu tablolar/routine'ler baştan yaratılmamalıdır.

## 15. Sıfır kalıntı kanıtı

Yalnızca grep sonucunu PR açıklamasına yapıştırma yeterli değildir.

- Go package/import graph, command entrypoint ve module dependency graph.
- SQL object inventory (tables, views, routines, triggers, policies, grants).
- Config key/env/secret producer-consumer graph.
- Generated artifact manifest ve generator testleri.
- Taskfile task invocation graph.
- Compose service/volume/secret/network graph.
- Docker build target/stage/artifact graph.
- CI/release job dependency graph.
- Docs/skills/AGENTS normatif referans grafiği.

Her eşleşmenin disposition'ı `REMOVED`, `REPLACED_BY_NEW_GOOSE`, `PRESERVED_NON_MIGRATION` veya `BLOCKED` olarak kaydedilir. `BLOCKED` veya sahibi/consumer'ı bilinmeyen tek bir eşleşme varken sıfır kalıntı iddiası yapılamaz.

## 16. CI ve kabul kapısı

PR CI'ına aşağıdaki yeni kapılar eklenir:

1. Ephemeral PostgreSQL 18 + pgvector üzerinde boş DB fresh initialize testi.
2. Aynı DB üzerinde ikinci initialize idempotency testi.
3. İki tenant için aynı Goose runner ve schema isolation testi.
4. API/worker DDL permission negative testleri.
5. Initialize ordering: Foundation fail → Goose/Application yok; Goose fail → Application yok.
6. Uygulanmış Goose migration checksum/immutability kontrolü: mevcut migration dosyası değiştirildiğinde CI fail eder. Bu kontrol Goose native version history yerine custom runtime state üretmez; Git tabanlı test/guard olabilir.
7. Eski executable/package/config/env/deployable isimleri için repository guard. Guard, domain kelimelerini körlemesine yasaklamaz; bilinen eski symbol/path/env allowlist'i ve açıklamalı disposition dosyasıyla çalışır.
8. `go test ./...`, `go vet ./...`, integration compile, `fileguard`, `archguard`.

Tam `task dev` komutu local release validation olarak çalışır; CI'da aynı INITIALIZE kod yolunu çağıran disposable compose/E2E job'u bulunur. Docker yoksa bu job başarıyla atlanamaz; runner Docker sağlayamıyorsa PR blocker olarak kalır.

## 17. Recovery ve veri güvenliği

- Eski local development volume veya kullanıcının mevcut database'i otomatik silinmez, migrate edilmez ya da adoption yapılmaz.
- Recovery backup restore, yeni schema lifecycle'ından ayrı bir operasyon olarak ele alınır; restore edilen DB'nin Goose history'si varsa native Goose history'dir, eski custom migration state'i kabul edilmez.
- Fresh rehydrate hedefinde önce boş hedef doğrulanır, sonra initialize çalışır, ardından recovery payload restore edilir. Restore öncesi Goose history yaratılmış DB üzerine dump restore etmenin çakışma ihtimali gerçek integration testle çözülür; güvenli sıra testle kanıtlanır ve bu iki modelden biri seçilir.
- `task local:destroy` yalnız açıkça istenen local volume resetidir; task dev/initialize tarafından çağrılmaz.
- Restore edilen database'e otomatik Goose up çalıştırılmaz. Restore compatibility ayrı release/ops kararıdır; bu task setinde upgrade yolu tasarlanmaz.

## 18. Nihai teslim ve kanıt
Tamamlanma raporu package/file disposition CSV/Markdown, SQL object disposition, migration-stream owner coverage, Taskfile/Compose dönüşüm tablosu, generated artifact farkı, fresh install/idempotency/tenant/permission test çıktıları ve eski referans taraması içermelidir. Sadece "grep temiz" veya "testler geçti" yeterli değildir.
