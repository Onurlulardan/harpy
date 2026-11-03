# 03 — GOOSE

## 1. Amaç

Goose, Harpy'nin bütün platform ve tenant schema değişiklikleri için tek ve mutlak migration otoritesidir.

Bu görev mevcut migration sistemini Goose'a taşımak değildir. Mevcut migration runtime'ı, SQL corpus'u, migration catalog'u, release/manifest zinciri, özel history modeli, tenant executor'ı ve bunların bütün bağımlılıkları silinir. Yeni Goose kullanımı boş bir zeminde oluşturulur.

Tek resmi implementation **github.com/pressly/goose/v3** olacaktır. Goose'un native discovery, execution, transaction ve version history davranışları dışında ikinci bir migration motoru veya policy katmanı yazılmaz.

## 2. Fresh schema kuralı

İlk kurulum yalnızca boş database üzerinde yapılır.

Yeni schema başlangıcı:

- Eski migration zincirinden türetilmez.
- Eski SQL migration dosyaları yeni klasöre kopyalanarak oluşturulmaz.
- Eski migration version numaraları korunmaz.
- Eski migration dosyalarının sıralaması/alias/supersession ilişkisi taşınmaz.
- Eski migration manifest veya catalog verisi kullanılmaz.
- Eski database'den Goose version tablosuna kayıt yazılmaz.
- Goose baseline/adoption/set-version işlemiyle var olan eski schema yeni sistemin kurulmuş hâli gibi gösterilmez.

Yeni initial schema, mevcut ürün gereksinimleri ve güncel domain contract'ları üzerinden yeniden modellenir. Bunun sonucunda gereken platform ve tenant tabloları, ilişkiler, index'ler, constraint'ler, sequence'ler, enum/type'lar, function'lar, trigger'lar ve view'lar yeni Goose SQL migration dosyaları olarak sıfırdan yazılır.

Yeni başlangıç migration'ları tek bir temiz başlangıç seti olacak şekilde tasarlanır. Eski migration corpus'undan dosya kurtarma veya yalnızca format değiştirme yapılmaz. Eski dosya geçmişi ve eski SQL gövdesi yeni migration source'una dahil edilmez.

## 3. Tek otorite ve kapsam

Goose aşağıdaki PostgreSQL schema nesnelerinin tamamını yönetir:

- Database içindeki platform schema nesneleri.
- Tenant schema nesnelerinin tamamı.
- Table ve partitioned table.
- Column, type ve default.
- Primary key, unique/check/exclusion constraint.
- Foreign key ve referential action.
- Index ve index değişiklikleri.
- Sequence ve identity davranışları.
- Enum, domain ve composite type.
- View ve materialized view.
- Function, procedure ve trigger.
- RLS policy ve ilgili schema güvenlik nesneleri.
- Extension sonrasında gereken schema etkili nesneler.
- Yeni uygulama modüllerinin ihtiyaç duyduğu domain tabloları.
- Sonradan yapılacak bütün schema değişiklikleri.

Foundation yalnız PostgreSQL altyapısını kurar. Application startup schema değiştirmez. Başka hiçbir Go package veya script schema DDL'i uygulayamaz.

## 4. Migration source düzeni

Platform ve tenant migration dosyaları kaynak kod içerisinde açık ve sabit bir düzende tutulur. Önerilen ayrım:

- Platform için tek native Goose source dizini.
- Tenant için tek native Goose source dizini.

Bu ayrım yalnız SQL kapsamını ve hedef schema'yı ayırır; ikinci bir runner/engine oluşturmaz. Her iki source aynı Goose implementation ve aynı çağrı yoluyla yürütülür.

Her source için:

- Dosyalar Goose'un native SQL migration formatında olur.
- Migration version/sıralama Goose native kurallarına bırakılır.
- Dosya isimleri açık, sıralanabilir ve çakışmasız olur.
- Goose'un desteklediği embed filesystem veya normal filesystem mekanizması kullanılır.
- Migration source'u custom manifest, allowlist, owner catalog veya path selection algoritmasıyla filtrelenmez.
- Bir migration dosyası başka custom parser tarafından okunmaz.
- Bir dosyanın uygulanıp uygulanmayacağı custom SQL veya Go state ile belirlenmez.

Repository'de eski SQL migration dosyaları yeni source ile aynı aktif klasörde bırakılamaz. Eski corpus tamamen silinir. Yeni source yalnız sıfırdan hazırlanmış migration'lardan oluşur.

## 5. Goose runner

Tek bir küçük runner/adapter bulunabilir. Görevi yalnızca:

1. Hedef PostgreSQL bağlantısını almak.
2. Platform veya tenant için doğrulanmış hedef schema'yı belirlemek.
3. Bağlantının search_path/schema kapsamını Goose native mekanizmasına doğru biçimde iletmek.
4. İlgili native migration source'unu Goose'a vermek.
5. İstenen standart Goose operasyonunu çağırmak.
6. Goose'un gerçek sonucunu ve hatasını aynen üst katmana döndürmek.

Runner aşağıdaki sorumluluklardan hiçbirini alamaz:

- Migration dosyası parse etme.
- Migration dosyalarını özel algoritmayla seçme/sıralama.
- SQL statement classification.
- Migration digest/hash üretme veya kontrol etme.
- Manifest/release çözümleme.
- Custom history/version tablosu okuma/yazma.
- Goose version tablosuna özel SQL uygulama.
- Baseline/adoption veya eski version aktarımı.
- Migration run/attempt/checkpoint/receipt yönetimi.
- Migration completion SQL function çağırma.
- Goose sonucunu custom generation/contract kurallarıyla doğrulama.
- Tenant'a özel farklı bir migration implementation çalıştırma.

Goose API/CLI seçenekleri dışında custom Goose store, custom migration state veya custom version source oluşturulmaz. Goose native PostgreSQL version history tek kayıttır.

## 6. Native version history

Goose'un native version history tablosu her hedef schema için Goose tarafından yönetilir.

- Tabloyu Foundation oluşturmaz.
- Uygulama oluşturmaz.
- Custom SQL script'i oluşturmaz.
- Go runner doğrudan INSERT/UPDATE/DELETE yapmaz.
- Başlangıçta eski version numarasıyla seed edilmez.
- Eski database'deki state buraya kopyalanmaz.
- Goose'un native migration execution sonucu dışında version değiştirilemez.

Platform ve her tenant schema için Goose'un native history davranışı korunur. Hedef schema/search_path yanlışsa işlem devam ettirilmez. Goose history'sinin yanlış schema'da oluşmadığı integration test ile kanıtlanır.

Goose native history tablosunun adı/şeması ve yapılandırması Goose'un desteklediği resmi API/CLI seçenekleriyle belirlenir. Custom version table implementasyonu veya custom store yazılamaz.

## 7. Platform ilk kurulum akışı

INITIALIZE, Foundation'ın başarılı olmasından sonra aynı Goose runner üzerinden platform migration source'unu çalıştırır.

Sıra:

1. Foundation sonucunun başarılı olduğunu doğrula.
2. Platform database bağlantısını Goose için hazırla.
3. Platform schema hedefini açıkça belirle.
4. Platform native Goose source'unu aç.
5. Goose native Up operasyonunu çalıştır.
6. Goose'un native history ve execution sonucunu kullan.
7. Hata varsa gerçek hatayı döndür ve Application aşamasını başlatma.
8. Başarı varsa APPLICATION aşamasına geç.

Platform migration'larının dışında ayrıca schema bootstrap SQL'i çalıştıran ikinci bir yol olamaz.

## 8. Tenant ilk kurulum ve otomatik provisioning

Otomatik tenant provisioning aşağıdaki kesin mimariyle yapılır. Bu, beşinci bir lifecycle aşaması veya sürekli çalışan executor değildir.

1. Harpy API tenant oluşturma isteğini alır. Tenant kaydını `PROVISIONING` durumuna getirir. API DDL yetkisi taşımaz.
2. API, güvenilir bir mekanizma üzerinden kısa ömürlü Tenant Provisioning Job başlatma isteği gönderir. Job image, command, service account, secret mount ve network ayarları kullanıcı girdisiyle değiştirilemez; sabit template'dir.
3. One-shot Tenant Provisioning Job aynı Harpy release image'ını ve aynı binary'yi kullanır. Aynı Goose Runner implementation'ını çağırır. Ayrı migrator binary veya ikinci Goose implementation yazılmaz.
4. Job, tenant ID'den güvenilir biçimde üretilen canonical schema adıyla hedef tenant schema'sını oluşturur (Foundation namespace/ownership/privilege kuralları içinde). Kullanıcı girdisi schema adını belirlemez.
5. Job, aynı Goose Runner ile yalnız o tenant'ın migration source'unu çalıştırır. Goose native history ilgili tenant schema içinde oluşur.
6. Migration başarılıysa tenant `ACTIVE` durumuna geçer ve kullanıma açılır. Başarısızsa `ACTIVE` olmaz ve hata gerçek sebebiyle raporlanır.
7. Job işlem bitince kapanır. Sürekli çalışan tenant migration executor, daemon, worker veya migration-specific queue yoktur.
8. Migration credential'ı API veya worker'a verilmez. Yalnız Job tarafından güvenilir secret mekanizmasından alınır ve işlem sonrası erişilemez.
9. Aynı tenant için eşzamanlı provisioning PostgreSQL advisory lock veya eşdeğer native mekanizma ile engellenir. Custom migration state, catalog, checkpoint veya receipt oluşturulmaz.
10. Tenant Job credential'ı yalnız hedef tenant schema'sında DDL yapabilir. Platform ve diğer tenant schema'larına erişemediği negative testlerle kanıtlanır.

One-shot Job yeni lifecycle aşaması değildir. INITIALIZE platformun ilk kurulumunu yapar; Job sonradan açılan tenant için aynı Foundation ve Goose kod yolunu çağırır.

## 9. Goose transaction ve eşzamanlılık

Her migration'ın PostgreSQL transaction gereksinimi açıkça belirlenir. Transaction dışı çalışması gereken DDL varsa Goose'un native migration yönergelerine uygun şekilde tanımlanır ve yarım kalma davranışı test edilir.

Aynı hedef schema üzerinde eşzamanlı Goose çağrılarının engellenmesi gerekiyorsa önce Goose'un native locking davranışı değerlendirilir. Yalnız native davranış yetersizse PostgreSQL advisory lock gibi geçici bir koordinasyon mekanizması kullanılabilir. Bu lock için özel state tablosu, lease tablosu, migration run kaydı veya custom history oluşturulmaz.

Lock yalnız eşzamanlı çalışmayı engeller; migration planı, migration state'i veya version bilgisini tutmaz.

## 10. Migration yazım standartları

Her yeni migration:

- Tek ve anlaşılır bir schema değişikliği amacı taşır.
- Gereksiz DDL/data işlemlerini aynı dosyaya yığmaz.
- Foreign key ve index bağımlılıklarını doğru sırada kurar.
- Gerekli constraint'leri açıkça tanımlar.
- Tenant schema kapsamını yanlışlıkla platform schema'sına taşırmaz.
- Platform migration'ında tenant verisine, tenant migration'ında başka tenant schema'sına dokunmaz.
- PostgreSQL identifier, reserved word, type ve locking davranışını dikkate alır.
- Gerekli geri alma davranışını Goose down migration kurallarıyla tanımlar; veri kaybına yol açabilecek down işlemleri gelişigüzel yazılmaz.
- Uygulama kodunun beklediği tablo/kolon/nullability/index/FK davranışlarıyla integration test edilir.

İlk kurulum için gerekli tüm schema nesneleri yeni initial migration setinde eksiksiz bulunur. Sonradan uygulanan schema değişiklikleri yeni forward migration olarak eklenir. Uygulanmış migration dosyası geriye dönük değiştirilmez.

## 11. Kesinlikle yasak olan eski mekanizmalar

Aşağıdaki kavramların hiçbirinin yeni implementation'ı oluşturulamaz ve eski implementation'ı repository'de tutulamaz:

- Custom migration catalog, manifest, bundle, archive veya selected migration listesi.
- Migration file hash/digest veya release hash.
- Release ID, release registry, release generation veya database/runtime generation.
- Migration run, attempt, receipt, checkpoint, requirement veya completion state.
- History alias, supersession, classification veya custom version mapping.
- Schema contract, binary schema contract veya custom schema readiness registry.
- Custom Goose store veya Goose history yerine kullanılan version table.
- Migration planner, selector, classifier, verifier, finalizer veya custom state machine.
- Sürekli çalışan tenant migration executor/migrator/worker/daemon/service veya migration-specific queue. Sabit template kullanan kısa ömürlü orchestrator Job izinlidir.
- Custom Goose completion routine veya migration-specific SQL function.
- Eski schema'yı Goose history'ye alma/baseline/adoption mekanizması.
- Eski migration dosyalarını yeni source'a dönüştüren/import eden generator.
- Goose'un önünde çalışan custom parser veya migration discovery sistemi.

Bu davranışlardan biri yeni bir isimle tekrar oluşturulursa kabul edilmez.

## 12. Test planı

Gerçek PostgreSQL üzerinde:

1. Tamamen boş database: platform migration'ları bütün gerekli schema'yı oluşturur.
2. Tamamen boş tenant schema: tenant migration'ları bütün gerekli tabloları oluşturur.
3. Tenant A ve Tenant B: her ikisi aynı runner/source ile kurulur.
4. History: Goose native version table yalnız hedeflenen schema'da oluşur.
5. Tekrar çağrı: Goose tamamlanmış migration'ları tekrar çalıştırmaz.
6. Hata: başarısız migration'ın alttaki PostgreSQL hatası korunur.
7. Retry: özel checkpoint olmadan Goose native history davranışıyla yeniden çağrı yapılır.
8. Tenant isolation: Tenant A migration'ı Tenant B schema'sını değiştirmez.
9. Permission: Goose yetkili rolü DDL yapabilir; API/worker rolü yapamaz.
10. Source: migration dosyaları Goose native discovery ile bulunur; custom selector yoktur.
11. Clean start: eski database veya eski migration state olmadan sıfırdan kurulum tamamlanır.
12. Eski corpus: eski migration SQL dosyası yeni source içinde bulunmaz.

## 13. Tamamlanma şartı

Goose, platform ve tenant schema DDL'inin tek otoritesidir. İlk schema sıfırdan yazılmış temiz migration source'larından kurulur. Goose native history dışında history/version/state yoktur. Eski migration sistemi Goose'a taşınmamış; tamamen silinmiştir. Runner yalnız resmi Goose implementation'ını çağırır.

## 14. Mevcut Goose/baseline kodunun yeniden yazım sınırı

Bugünkü `core/internal/platform/baseline/apply.go` resmi Goose Provider çağırıyor olsa bile yeni hedefin kabul edilebilir Goose implementation'ı değildir. Bu dosya özel `fstest.MapFS` snapshot seçimi, sabit legacy version listesi, ardışık Catalog-removal ve tenant-executor completion migration uygulaması içerir. `core/internal/platform/migrator/build_goose_provider.go` ve `build_tenant_goose_provider.go` da mevcut custom release/catalog/contract akışının parçasıdır. Bu mekanizma Goose kullanıyor diye korunamaz.

Yeni implementation yalnız resmi Goose API'sini çağırır. Başlangıç SQL'i normal Goose native source dizinlerinde bulunur; dosya adları Goose tarafından keşfedilir. `ApplyPlatform`, `ApplyTenant`, `applyPlatformTenantExecutor*`, `combineBaselineResults`, `AppliedVersions`, `fstest.MapFS` tabanlı özel baseline sağlayıcısı veya bunların eşdeğeri yeni tasarımda bulunmaz.

Mevcut dosyaların zorunlu disposition'ı:

- `core/internal/platform/baseline/`: tüm eski özel baseline Go/test kodu kaldırılır; yeni minimal Goose çağrı sınırı gerekiyorsa yeni package'ta sıfırdan yazılır.
- `core/internal/platform/baseline/sql/*.fresh.sql`: yeni source'a kopyalanmaz. Her birinin çıkardığı schema nesnesi disposition envanterine alınır.
- `core/internal/platform/migrator/`: Goose dışında kalan planner/selector/release/catalog/checkpoint/evidence/runner/tenant executor kodları silinir.
- `core/internal/platform/migration/`: custom migration document, release ve source abstraction'ları silinir.
- `core/internal/platform/migratoroutput/`: custom plan/up/verify çıktı modelleri silinir.
- `core/internal/platform/tenantmigrationexecutor/` ve `core/internal/platform/migrator/tenantmigrationexecutor/`: silinir; ayrı tenant process yoktur.
- `core/internal/platform/schemacontract/` ve `core/internal/platform/binarycontract/`: migration/release contract bağımlılıkları kaldırılır. Uygulama runtime readiness'ine gereken non-DDL bağlantı kontrolleri ayrı olarak tanımlanır.

### 14.1 Sekiz mevcut snapshot'ın tek tek ele alınması

Önceki envanterde görülen dosyalar:

| Snapshot | Disposition |
| --- | --- |
| `20261028000000_platform_baseline.fresh.sql` | Tarihsel platform schema envanteri; SQL gövdesi taşınmaz. |
| `20261028000001_tenant_baseline.fresh.sql` | Tarihsel tenant schema envanteri; SQL gövdesi taşınmaz. |
| `20261102000000_platform_catalog_removal_baseline.fresh.sql` | Catalog/migration framework cleanup; yeni DB'de tekrar üretilmez. |
| `20261102000001_tenant_catalog_removal_baseline.fresh.sql` | Tenant Catalog cleanup; yeni DB'de tekrar üretilmez. |
| `20261103000003_platform_tenant_executor_baseline.fresh.sql` | Tenant executor runtime rutinleri; yeni modele taşınmaz. |
| `20261103000004_platform_tenant_executor_completion_baseline.fresh.sql` | Legacy checkpoint completion/reopen rutinleri; yeni modele taşınmaz. |
| `20261103000005_platform_tenant_executor_claim_clock_baseline.fresh.sql` | Legacy tenant executor claim/clock düzeltmesi; yeni modele taşınmaz. |
| `20261103000006_platform_tenant_executor_completion_contract_baseline.fresh.sql` | Legacy executor completion/onboarding state transition; yeni modele taşınmaz. |

Bu listedeki bir SQL nesnesinin güncel bir application davranışını sağladığı görülürse yalnız o davranış ilgili domain owner altında, migration dışı bir modelle yeniden tasarlanır. Legacy function adı veya signature'ı korunarak yeni Goose'a kopyalanmaz.

### 14.2 Eski stream ve snapshot'tan schema tasarımına geçiş

`contracts/storage/migration-streams.yaml` ile include ettiği 14 owner JSON dosyası yalnız tarihsel kapsam kontrolüdür. Önceki sayımdaki 312 stream bir kabul kontrol toplamıdır. Her owner ve stream için üretilen DB nesneleri güncel Go SQL consumers, storage owner contracts ve API/domain behavior ile çapraz doğrulanır. Her nesne `KEEP_PRODUCT_SCHEMA`, `RECREATE_AS_NEW_GOOSE_DDL`, `REMOVE_LEGACY_STATE` veya `UNRESOLVED` olarak sınıflandırılır.

Her `RECREATE_AS_NEW_GOOSE_DDL` satırında yeni dosya/satır, owner, consumer ve PostgreSQL integration test adı bulunur. Eski stream version, SQL gövdesi, supersession ilişkisi veya selected order yeni source'a taşınamaz. 14 owner'dan biri veya herhangi bir nesne `UNRESOLVED` kalırsa task başarısızdır.

### 14.3 Goose immutability

Yeni migration source için Git tabanlı bir immutability gate ekle:
- PR'da değiştirilen/değiştirilmeye çalışılan mevcut migration dosyalarının path listesi çıkarılır.
- Default/base branch'te zaten mevcut olan migration dosyasında içerik değişikliği CI'ı fail eder.
- Yeni migration eklenmesine izin verilir.
- Bu kontrol Goose native history'yi değiştirmez ve runtime custom digest/history tablosu üretmez.
- İstisna/allowlist ile eski migration dosyasını değiştirmek mümkün kılınmaz; düzeltme yeni forward migration ile yapılır.

## 15. Tenant provisioning Job ve legacy schema-executor temizliği

Otomatik tenant provisioning §8'deki one-shot Tenant Provisioning Job ile yapılır. Job aynı Harpy image/binary'sini ve aynı Goose Runner'ı kullanır. Sürekli çalışan executor/worker/daemon/queue veya ayrı migrator binary oluşturulmaz.

`core/internal/platform/migrator/schemaexecutor/`, typed DDL IR ve bunlara bağlı tüm command, queue, config, secret, contract, test ve deployment referansları ForgeCRUD kaynaklı kalıntıdır. Tamamen kaldırılır. Harpy'de dynamic form, form publish, form metadata, record CRUD veya benzeri ürün davranışı yoktur. JSONB/EAV tasarlanmaz; form davranışı için acceptance test yazılmaz; yerine yeni form/config alt sistemi kurulmaz. Tüm consumer ve deployment bağlantıları sıfır kalıntı denetimine dahildir.

### 15.1 Tenant schema history

Tenant A ve Tenant B paralel Job çağrılarında Goose native history/DDL çapraz schema'ya yazılmaz. Custom tenant version table veya executor coordination state eklenmez.

### 15.2 Tamamlanma

Fresh PostgreSQL platform/tenant testleri, tenant Job privilege/isolation testleri geçmeli; schema-executor/typed DDL IR consumer'ları kaldırılmalı; immutable migration gate çalışmalı ve historical stream/snapshot disposition UNRESOLVED içermemelidir.
