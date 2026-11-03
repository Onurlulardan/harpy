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

## 8. Tenant ilk kurulum ve provisioning

Her tenant, yeni ve boş tenant schema'sı olarak kurulur. Eski tenant schema'sı veya eski tenant migration state'i yeni sistemin girdisi değildir.

Tenant akışı:

1. Tenant ID mevcut tenancy domain kurallarıyla doğrulanır.
2. Canonical tenant schema adı üretilir ve identifier güvenliği doğrulanır.
3. Foundation sorumluluğundaki namespace/ownership/privilege işlemi uygulanır.
4. Aynı Goose runner tenant migration source'unu bu schema için çalıştırır.
5. Goose native history ilgili tenant schema içinde oluşur.
6. Tüm migration'lar başarılıysa tenant mevcut application provisioning akışında aktif edilir.
7. Goose başarısızsa tenant aktif edilmez ve hata gerçek sebebiyle raporlanır.

Platform ve tenant için ayrı migrator servisi, ayrı binary, ayrı runner, tenant executor, migration worker veya migration-specific queue oluşturulmaz. Tenant işlemi aynı kod yolu, aynı Goose implementation ve aynı native history mekanizmasıyla yapılır.

Tenant provisioning sırasında admin yetkisinin yönetimi açıkça tanımlanır. API/worker runtime credential'ı DDL yetkisi almaz. Ayrı bir deployable ekleyerek bu yetki sorunu çözülmez.

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
- Tenant migration executor/migrator/worker/daemon/service.
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