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

Platform ve tenant namespace'leri açıkça belirlenir. Schema isimleri canonical tenant identity kuralından üretilir; kullanıcı girdisi, URL slug'ı veya HTTP header doğrudan SQL identifier olarak kullanılmaz.

Tenant schema oluşturma işlemi:

1. Tenant kimliğinin mevcut tenancy domain kuralıyla doğrulanması.
2. Schema adının tek ve deterministik canonical mapping ile hesaplanması.
3. Identifier'ın PostgreSQL limitleri ve karakter kurallarına uygun olduğunun doğrulanması.
4. Schema'nın Foundation yetkisiyle oluşturulması.
5. Ownership ve default privilege kurallarının atanması.
6. Goose'un aynı native runner ile doğru schema/search_path hedeflenerek çağrılması.
7. Goose başarılı olmadan tenant'ın aktif edilmemesi.

Tenant provisioning için ayrı migration engine, tenant executor, migrator process, worker veya özel history kaydı oluşturulmaz. Foundation yalnız namespace ve privilege işini yapar; tenant tablolarını Goose kurar.

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