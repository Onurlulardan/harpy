# 02 — DATABASE FOUNDATION

## 1. Amaç

DATABASE FOUNDATION, yeni Harpy kurulumunun kullanacağı boş PostgreSQL database'ini ve yalnızca database seviyesinde zorunlu altyapıyı hazırlar.

Foundation sıfırdan yazılır. Eski migrator, bootstrap veya schema preparation package'larının refactor edilmiş hâli olamaz. Eski package, struct, SQL routine veya custom state modelini yeniden kullanamaz.

Foundation bir migration engine değildir. Domain schema oluşturmaz ve schema version/history tutmaz.

## 2. Sorumluluk sınırı

Foundation yalnız Goose ve Application çalışmadan önce PostgreSQL seviyesinde hazır olması gereken nesneleri kurar:

- Yeni, boş database'in oluşturulması veya önceden ayrılmış boş database'e bağlanılması.
- Gerekli platform schema namespace'inin oluşturulması.
- Tenant schema'ları oluşturulurken kullanılacak temel namespace/ownership/privilege kurallarının hazırlanması.
- Ayrıcalıklı bootstrap/owner rolü ve sınırlı runtime rollerinin oluşturulması.
- Zorunlu PostgreSQL extension'larının kurulması.
- Database ve schema ownership'inin atanması.
- Goose migration bağlantısının gerekli DDL yetkilerinin verilmesi.
- API ve worker rollerinin minimum runtime yetkilerinin verilmesi.
- GRANT ve ALTER DEFAULT PRIVILEGES ayarlarının kurulması.
- Bağlantı, search_path ve public schema güvenliğinin ayarlanması.

Bunların dışında bir iş Foundation'a eklenemez.

## 3. Oluşturulmayacak nesneler

Foundation aşağıdakilerin hiçbirini oluşturamaz:

- Platform/domain/business tabloları.
- Tenant business tabloları.
- Application ayar, kullanıcı, rol, permission, workflow veya modül tabloları.
- Goose version history tablosu.
- Custom migration version/history/catalog tablosu.
- Installation state, bootstrap state veya generation tablosu.
- Migration run/attempt/checkpoint/receipt/requirement tabloları.
- Migration release/manifest/digest kaydı.
- Schema contract veya migration readiness tablosu.
- Tenant migration state tablosu.
- Goose tamamlanmasını işaretleyen custom SQL function/procedure.

**goose_db_version tablosunu yalnızca Goose kendi native davranışıyla oluşturur ve yönetir.** Foundation bu tabloya DDL uygulayamaz, elle INSERT/UPDATE/DELETE yapamaz veya önceden seed edemez.

Foundation tarafından kurulacak bootstrap tabloları varsayılan olarak sıfırdır. Gerçek PostgreSQL davranışı açısından zorunlu bir tablo ihtiyacı kanıtlanmadığı sürece tablo oluşturulmaz. Goose native history bu istisna değildir; Goose'a aittir.

## 4. Boş database ön koşulu

Foundation yalnızca fresh kurulum için ayrılmış boş database üzerinde çalışır.

Başlamadan önce:

1. Bağlantı hedefini doğrula.
2. Database adını/kimliğini doğrula.
3. Beklenmeyen uygulama tabloları, eski framework metadata'sı veya domain nesnesi bulunmadığını kontrol et.
4. Database yeni kurulum için temiz değilse hiçbir DDL çalıştırmadan fail closed dön.

Foundation eski database'i temizlemeye, eski schema'yı yeniden oluşturmaya, eski tabloları silmeye, eski role'leri dönüştürmeye veya eski metadata'yı Goose'a taşımaya çalışmaz.

Otomatik DROP DATABASE, DROP SCHEMA CASCADE, TRUNCATE veya genel wildcard cleanup yasaktır. Yanlış database'e bağlanma ihtimalinde veri silmek yerine kurulum durdurulur.

## 5. Role ve yetki modeli

Yeni rol tasarımı mümkün olan en az yetki prensibiyle sıfırdan tanımlanır. Rol isimleri ve sorumlulukları açık olmalıdır; eski migration framework'ünün role modelini yalnız isim değiştirerek taşımak yasaktır.

Asgari sorumluluk ayrımı:

### Bootstrap/owner erişimi

- INITIALIZE ve Foundation işlemlerinde kullanılır.
- Database, schema, role, extension, ownership ve grant işlemlerini yapabilecek yetkiye sahiptir.
- API/worker process'lerine verilmez.
- Log veya generated artifact içine yazılmaz.
- Uygulama runtime'ı tarafından okunamaz.

### Goose migration erişimi

- Yalnız Goose çağrısında kullanılır.
- Platform/tenant schema DDL işlemlerini yapabilecek yetkiye sahiptir.
- Goose işlemi bittiğinde uygulama process'ine aktarılmaz.
- API/worker connection pool'u tarafından kullanılamaz.
- Gereksiz cluster-wide superuser yetkisi verilmez; PostgreSQL'in izin modeli içinde gerekli en dar yetki tanımlanır.

### Application runtime erişimi

- API ve worker tarafından kullanılır.
- İhtiyaç duyduğu schema'larda SELECT/INSERT/UPDATE/DELETE ve gerekli sequence/function execute izinlerine sahiptir.
- CREATE/ALTER/DROP schema/table/function/type/extension veya role yönetme yetkisi yoktur.
- Başka tenant schema'larına erişemez.
- Goose history tablosunu değiştiremez.

Aynı credential'ın hem Goose hem API/worker tarafından kullanılması kabul edilmez.

## 6. Schema ve tenant namespace kuralları

Tenant schema kurulumu APPLICATION/tenancy isteğiyle orchestrator tarafından başlatılan kısa ömürlü one-shot Job üzerinden yapılır. Job aynı Harpy image digest'ini, aynı binary'yi ve tek Goose Runner'ı kullanır; ayrı migrator binary/implementation, sürekli executor/daemon/worker veya migration queue oluşturulmaz.

API yalnız sabit trusted Job template'ini tenant ID ile tetikleyebilir. Image, command, service account, secret mount, volume ve network ayarları kullanıcı girdisiyle değiştirilemez; orchestrator policy/admission ile sabitlenir. Job Foundation tenant namespace/ownership işlemini ve aynı Goose Runner ile yalnız hedef tenant migration source'unu çalıştırır.

DDL credential kısa ömürlü secret injection ile yalnız Job'a verilir. API/worker credential'ı okuyamaz veya DDL yapamaz. Tenant Job credential'ı yalnız hedef tenant schema'sında DDL yapabilir; platform ve diğer tenant schema'larına erişemediği PostgreSQL negative testleriyle kanıtlanır. Cluster-wide superuser/CREATEDB/CREATEROLE verilmez.

Tenant domain kaydı işlem başında PROVISIONING, Job/Goose başarısında ACTIVE olur; hata durumunda aktif edilmez. Bu gerçek tenancy lifecycle state'idir, migration attempt/checkpoint/receipt değildir. Retry/concurrency custom migration state'i olmadan PostgreSQL/orchestrator native davranışıyla yönetilir. Orchestrator bu izolasyonu sağlayamıyorsa API'ye DDL yetkisi verilerek fallback yapılmaz; provisioning fail closed olur.

Foundation namespace/ownership/default privilege işlemini yapar; Goose tenant tablolarını kurar. API/worker DDL credential taşımaz.

## 7. Extension yönetimi

Yeni kurulumun çalışması için zorunlu extension'lar açık bir kaynak listesinde tanımlanır. Extension seçimi eski manifest/digest sisteminden alınmaz.

Foundation:

- Gerekli extension'ların PostgreSQL ortamında desteklendiğini doğrular.
- Eksik zorunlu extension'ı kurar veya kurulamıyorsa açıklayıcı hata verir.
- Extension kurulumu için gereken ayrıcalığı yalnız bootstrap aşamasında kullanır.
- Extension kurulum sonucunu custom version registry veya digest tablosuna yazmaz.

Extension sonrasında schema nesneleri veya fonksiyonlar üzerinde gereken bütün değişiklikler Goose SQL migration'larının sorumluluğudur. Foundation extension version history tutmaz.

## 8. Implementation kuralları

Yeni Foundation implementation'ı:

- Küçük, açık ve test edilebilir PostgreSQL işlemlerinden oluşur.
- İstenen nesnenin varlığını ve beklenen owner/privilege durumunu doğrular.
- Tekrar çağrıldığında mevcut doğru nesneyi korur.
- Farklı/şüpheli nesne gördüğünde sessizce üzerine yazmaz; açık hata üretir.
- Bir işlemin kısmen tamamlanması halinde tekrar denenebilir olur.
- Transaction kullanılabilen işlemleri transaction içinde yürütür.
- Transaction dışı PostgreSQL işlemlerinin yarım kalma davranışını açıkça ele alır.
- Password ve secret değerlerini hata/log mesajlarına dahil etmez.
- Schema DDL'ini veya domain tablo DDL'ini Go source içine gömerek dağıtmaz.

Foundation için bağımsız migration discovery, SQL corpus, manifest veya custom executor oluşturulmaz. Foundation SQL'i yalnız DB altyapı işlemlerini yapar.

## 9. Test planı

Gerçek desteklenen PostgreSQL sürümünde aşağıdaki testleri çalıştır:

1. Boş database için database/role/schema/extension/ownership/grant kurulumu.
2. Foundation ikinci kez çalıştırıldığında duplicate role/schema/extension oluşmaması.
3. Var olan doğru owner/grant'in gereksiz yere bozulmaması.
4. Yanlış owner veya beklenmeyen privilege durumunda açık hata üretilmesi.
5. Boş olmayan/eski database'e bağlanıldığında hiçbir nesnenin silinmemesi ve işlemin durması.
6. Foundation sonrasında business/domain tablosu bulunmaması.
7. Goose history tablosunun Foundation sonrasında bulunmaması.
8. Goose rolünün gerekli DDL işlemlerini yapabilmesi.
9. API/worker rolünün CREATE/ALTER/DROP yapamaması.
10. Bir tenant rolünün başka tenant schema'sına erişememesi.
11. Secret değerlerinin loglarda görünmemesi.
12. Tenant schema adının canonical ve güvenli şekilde üretildiği.

## 10. Tamamlanma şartı

Foundation yalnızca boş PostgreSQL database'inin altyapısını kurar. Domain tabloları, migration history veya custom state üretmez. Goose native history'sine hiçbir şekilde dokunmaz. Eski Foundation/migrator kodu ve ona bağlı config, SQL, test, CLI ve deployable kalıntıları silinmiştir.

## 11. Mevcut PostgreSQL foundation kaynaklarının disposition'ı

Bugünkü implementation `core/internal/platform/baseline/apply.go` içindeki özel baseline uygulayıcısını ve `core/internal/platform/postgres/foundation_baseline.go` bağlantısını içerir. Bunlar yeni Foundation veya Goose runner olarak korunamaz. Her biri sıfırdan tasarlanan yeni sınırda karşılanır.

Kontrol edilecek mevcut package ve kaynaklar:

- `core/internal/platform/baseline/` ve altındaki bütün Go/test/SQL dosyaları.
- `core/internal/platform/postgres/foundation_baseline.go` ve `foundation_baseline_test.go`.
- `core/internal/platform/migrator/` içindeki database preflight, extension evidence, owner/grant, foundation evidence ve baseline orchestration consumer'ları.
- `core/internal/composition/migrator/` içindeki startup/bind/config wiring.
- `contracts/storage/owners/platform_migrator.json` ve bu owner kaydını yükleyen/generating kod.
- `ops/local/compose.tools.yml`, `ops/local/compose.lifecycle.yml`, `ops/local/compose.inventory.yml` ve `.taskfiles/local*.yml` içindeki secret, volume ve service tanımları.

Foundation'a taşınabilecek tek sorumluluk PostgreSQL database/role/schema/extension/ownership/grant/default-privilege hazırlığıdır. Eski `environment_identity`, extension registry/evidence, foundation receipt/finalization, migration readiness, release generation veya manifest verileri isim değiştirerek korunamaz.

### 11.1 PostgreSQL nesne allowlist'i

Foundation'ın kurabileceği nesneler implementation öncesi bir allowlist olarak kod/test ile sabitlenir:

- Database bağlantısı ve database-level güvenlik ayarları.
- Açıkça listelenmiş login/no-login roller ve üyelikleri.
- Platform namespace ve tenant schema namespace/ownership kuralları.
- Gereken extension kurulumları.
- Schema/database grants ve default privileges.

Bu allowlist'in dışında tablo, view, function, procedure, trigger, type, sequence, policy veya Goose history oluşturulamaz. Foundation integration test'i katalog sorgularıyla yalnız allowlist nesnelerinin oluştuğunu kanıtlar. Goose sonrasında domain schema nesneleri yalnız Goose source tarafından gelir.

### 11.2 Secret ve yetki sınırı

Eski `migrator.env`, `schema-executor.env`, `tenant-migration-executor.env` ve local-release database secret üretimleri yeni Foundation credential modeline alias edilmez. Yeni ayrıcalıklı credential yalnız initialize process'ine verilir; API/worker environment, mounted secret volume veya artifact setinde bulunmadığı negatif testle gösterilir. Tenant-scoped DDL secret yalnız orchestrator/secret manager üzerinden one-shot provisioning Job'a verilir. Log ve hata mesajlarında DSN/password/token/private key redaction testleri bulunur.

### 11.3 Tamamlanma

Foundation task'ı tamamlanmış sayılmaz; gerçek PostgreSQL 18 disposable testinde boş database preflight, role/schema/extension/grant kurulumu, ikinci çağrıda idempotency, dolu/eski DB'de sıfır DDL, Goose history yokluğu ve API/worker DDL yetkisizliği kanıtlanana kadar.