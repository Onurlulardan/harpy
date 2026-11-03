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