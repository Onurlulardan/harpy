# 02 — DATABASE FOUNDATION

## Amaç

PostgreSQL'i ve uygulamanın çalışması için zorunlu altyapıyı tek seferlik bootstrap ile hazırla. Foundation bir migration engine değildir ve kendi schema evolution mekanizmasını kuramaz.

Foundation'ın işi PostgreSQL katmanını hazırlamaktır:

- PostgreSQL database'inin oluşturulması veya mevcut database'e güvenli bağlanılması
- Gerekli platform ve tenant schema namespace'lerinin oluşturulması
- Gerekli PostgreSQL role'lerinin oluşturulması
- İlk kurulumda zorunlu extension'ların yüklenmesi
- Schema/table/function ownership'in doğru role'lere verilmesi
- Runtime, owner ve bootstrap erişimlerinin ayrılması
- GRANT ve ALTER DEFAULT PRIVILEGES altyapısının kurulması
- Goose ve Application'ın ihtiyaç duyduğu en küçük bağlantı/izin temelinin hazırlanması

## Foundation sınırı

Foundation yalnızca DB seviyesinde zorunlu olan, kurulumu Goose'tan önce yapılması gereken nesneleri oluşturur.

- Foundation, business module tablolarını, tenant business tablolarını veya zamanla değişecek platform domain tablolarını oluşturmaz. Bunların ilk oluşturulması da Goose migration'ına aittir.
- Goose çalışmadan önce var olması teknik olarak zorunlu bir bootstrap tablosu varsa bunun istisna olduğu belgelenir. Bu tablo bir migration history veya custom initialization registry olamaz.
- Goose'un goose_db_version tablosu Foundation tarafından yaratılmaz, seed edilmez, güncellenmez veya özel SQL ile yönetilmez. Goose native mekanizması kendi history tablosunun sahibidir.
- Foundation tabloları dahil Foundation tarafından oluşturulan herhangi bir nesnenin sonraki schema değişikliği yalnızca Goose ile yapılır.
- Foundation kodunda domain/business tablo DDL'i bulunamaz.
- Foundation çalışmasının tekrarı mevcut database, schema, role, extension, owner veya grant'leri silmez.

## Role ve permission modeli

- Privileged bootstrap credential yalnız INITIALIZE / DATABASE FOUNDATION sırasında kullanılır.
- Goose DDL çalıştırırken gerekli owner/migrator database yetkisini kullanır; bu, ayrı bir daemon veya servis oluşturma gerekçesi değildir.
- API ve worker runtime credential'ları DDL, schema create/drop, extension admin veya role yönetme yetkisine sahip olamaz.
- Owner role ile runtime role birbirinden ayrılır.
- Schema owner'ları ve default privileges açıkça tanımlanır; yeni Goose nesneleri beklenen owner'a ait olur.
- Tenant schema'sında runtime erişim yalnız doğrulanmış tenant scope ve beklenen role izinleriyle sağlanır.
- Public schema'ya gereksiz CREATE/TEMP yetkisi verilmez.
- Secret değerleri log, error, task output veya generated config içine yazılmaz.

Mevcut Harpy role isimleri ve security modelini önce incele; gereksiz yeni role ailesi icat etme. Mevcut roller korunacaksa yalnız Foundation'ın tanımladığı DB sorumluluklarını karşılayacak şekilde sadeleştir.

## Extension yönetimi

- İlk kurulumda zorunlu extension'lar Foundation tarafından kurulur.
- Extension yoksa açık hata verilir veya Foundation'ın açıkça tanımladığı şekilde kurulur.
- Extension sürüm upgrade'i veya sonradan eklenecek schema etkili extension değişikliği Goose migration'ı dışında yapılamaz.
- Extension manifest/digest veya release attestation oluşturulmaz.
- Extension durumunu ikinci bir schema version registry olarak saklama.

## Tenant schema hazırlığı

Tenant schema adı client slug veya request header'dan doğrudan türetilmez. Mevcut canonical TenantID/schema-name kuralları kullanılır.

Yeni tenant için:

1. Foundation yetkisiyle yalnız boş tenant schema namespace'i oluştur.
2. Ownership ve gerekli privilege'ları ata.
3. Goose'un native tenant migration dizinini aynı Goose runner ile çalıştır.
4. Goose başarılı olmadan tenant'ı aktif veya API tarafından kullanılabilir hale getirme.

Tenant schema'sını create etmek için tenant migration executor servisi, migration worker'ı, migration-specific table veya özel version ledger oluşturma.

## Mevcut kurulumların korunması

- Var olan database, tenant schema ve business tabloları yeniden oluşturulamaz.
- Role/owner düzeltmesi gerektiğinde yalnız hedef nesnede, geri alınabilir ve doğrulanabilir işlem uygula.
- Foundation mevcut nesnenin içeriğini temizleyemez veya tablolara truncate uygulayamaz.
- Mevcut custom migration metadata tabloları Foundation'da korunmaz. Bunların kaldırılması, Goose çalışır hale geldikten sonra Goose forward cleanup migration'ı ile yapılır.
- Eski framework'e ait metadata tablolarını business table sanıp koruma; fakat business data tablolarını migration metadata diye silme. Her nesnenin ownership ve consumer ilişkisini doğrula.

## Kesin yasaklar

- Foundation migration catalog, version table, applied migration listesi veya custom baseline state tutamaz.
- Foundation Goose history tablosuna doğrudan INSERT/UPDATE/DELETE yapamaz.
- Foundation domain/business table DDL'ini runtime Go koduna, init SQL script'ine veya entrypoint shell'ine dağıtamaz.
- Foundation migration selection, manifest parsing, schema contract verification veya release validation yapamaz.
- Foundation başarısız olduğunda database'i drop/recreate edemez.
- Uygulama credential'larına DDL privilege veremez.

## Test ve kabul

Gerçek desteklenen PostgreSQL sürümü üzerinde doğrula:

1. Tamamen boş ortamda database/role/extension/schema/owner/grant kurulumu başarılı.
2. Foundation ikinci kez çalıştırıldığında idempotent ve veri koruyucu.
3. Goose'un oluşturduğu tablolar doğru owner ve default privilege'lere sahip.
4. API/worker role'ü ile CREATE/ALTER/DROP TABLE, CREATE SCHEMA, CREATE EXTENSION ve role yönetimi başarısız.
5. Goose owner/migrator role'ü yalnız beklenen schema'larda gerekli DDL'i yapabiliyor.
6. Tenant A ve Tenant B schema sahipliği ve erişimi birbirinden ayrı.
7. Mevcut database üzerinde Foundation tekrar çalıştırıldığında business satırlarının kimliği ve sayısı korunuyor.
8. Goose history tablosu yalnız Goose tarafından oluşturulmuş ve yönetiliyor.
9. Foundation source ve script'lerinde uygulama/domain DDL'i bulunmuyor.

## Tamamlanma şartı

Foundation yalnız PostgreSQL altyapısını hazırlar. Schema evolution yapmaz, migration geçmişi tutmaz, Goose native history'ye dokunmaz ve uygulama tablolarını paralel biçimde yönetmez. Foundation sonrasında tüm schema değişikliklerinin tek sahibi Goose'tur.
