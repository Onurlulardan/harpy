# Harpy — Tek Veritabanı Yaşam Döngüsü Task Seti

## Bağlayıcı hedef

Harpy'nin kurulum ve veritabanı yaşam döngüsü **yalnızca ve tam olarak dört aşamadan** oluşacaktır:

1. **INITIALIZE**
2. **DATABASE FOUNDATION**
3. **GOOSE**
4. **APPLICATION**

Bunların dışında beşinci bir aşama, ayrı migration servisi, paralel schema lifecycle, özel migration runtime'ı veya alternatif migration otoritesi oluşturulmayacaktır. Bu dosyaların numaraları yalnızca iş sırasını belirtir; yeni runtime aşamaları tanımlamaz.

Bu task seti, repository'deki mevcut kurulum/migration mimarisine dair çelişen eski dokümanlardan ve skill hükümlerinden üstündür. Eski tasarımı korumak, adını değiştirmek, yeni bir abstraction altına saklamak veya uyumluluk amacıyla yaşatmak kabul edilmez.

## Nihai akış

INITIALIZE
→ DATABASE FOUNDATION
→ GOOSE
→ APPLICATION

İlk üç aşama başarıyla tamamlanmadan API, worker veya başka bir uygulama servisi başlatılamaz. Uygulama startup'ı schema DDL çalıştıramaz. Goose tamamlanmış schema'yı kullanır; schema'nın sahibi veya ikinci bir migration yöneticisi olmaz.

## Dört aşamanın kesin sınırları

### 1. INITIALIZE

İlk kurulumun tek giriş noktasıdır. Sıralamayı yönetir ve başarısız aşamada akışı durdurur. İlk kurulum dışında her deploy'da özel bir migration lifecycle başlatmaz. İlk kurulumun yarıda kesilmesi halinde tekrar çağrılabilmesi için güvenli ve idempotent olmalıdır. Bu idempotency özel bir initialization/version registry ile değil, gerçek PostgreSQL nesnelerinin ve Goose native history'sinin kontrol edilmesiyle sağlanır.

### 2. DATABASE FOUNDATION

PostgreSQL database'ini ve Goose/Application çalışması için zorunlu altyapıyı hazırlar: gerekli schema'lar, roller, extension'lar, ownership, grant/default privilege ve gerçekten kaçınılmaz bootstrap tabloları.

Foundation kendi schema version'ını, migration list'ini, release bilgisini veya geçmiş kaydını tutamaz. Goose'un kendi **goose_db_version** tablosunu Foundation oluşturmaz, düzenlemez veya taklit etmez; bu tablo Goose'a aittir.

Foundation aşamasında oluşturulması zorunlu olmayan bütün uygulama/platform/tenant tabloları ilk Goose migration'ında oluşturulur. Foundation'ın oluşturduğu sabit bir bootstrap tablosu varsa yalnızca Goose çalışmadan önce var olması teknik olarak zorunlu olduğu için bulunabilir; bu tablonun sonraki tüm DDL değişiklikleri de Goose migration'ı ile yapılır.

### 3. GOOSE

Platform schema'ları ve her tenant schema'sı dahil, tüm schema lifecycle'ın tek otoritesidir. Bundan sonraki bütün DDL değişiklikleri yalnızca Goose migration dosyalarıyla yapılır:

- Table, column ve data type
- Index, constraint, foreign key ve sequence
- Function, procedure ve trigger
- Schema, view ve diğer PostgreSQL schema nesneleri
- Tenant schema nesnelerinin tamamı

Migration geçmişi yalnızca Goose'un native version store'unda tutulur. Her hedef schema için Goose'un native history mekanizması kullanılır. Goose yanında ikinci bir geçmiş, catalog, manifest veya özel checkpoint bulunamaz.

Goose, sürekli çalışan ayrı bir servis değildir. INITIALIZE akışı ile tenant oluşturma akışında gereken yerde çağrılan tek seferlik Goose işlevidir. Ayrı bir migrator daemon/container/deployable veya tenant migration executor kurulmayacaktır.

### 4. APPLICATION

API, worker ve diğer uygulama servisleridir. Yalnızca Foundation ve Goose başarılı olduktan sonra başlarlar. Normal uygulama runtime credential'ları DDL yetkisi taşımaz. API/worker startup'ında migration, baseline, schema repair veya otomatik DDL çalıştırılmaz.

Yeni tenant oluşturulurken de aynı sıra uygulanır: gerekli tenant schema'sı Foundation yetkisiyle oluşturulur, aynı Goose runner tenant schema'sına migration uygular, başarıdan sonra tenant uygulamaya açılır. Bunun için ikinci bir migration sistemi, migrator process'i, manifest veya migration-specific job state oluşturulmaz.

## Kesin yasaklar

Aşağıdakilerden hiçbiri mevcut veya yeni bir isimle repository'de kalamaz:

- Custom migration catalog, migration manifest, selected migration listesi veya özel dosya seçme/sıralama algoritması.
- Migration digest/hash, migration release, release archive, release registry, database/runtime generation veya bunların attestation/receipt modelleri.
- Goose version geçmişine paralel custom version/history tablosu, custom Goose store veya Goose history'sini saran ikinci bir kayıt sistemi.
- Migration run, migration attempt, migration checkpoint, requirement catalog, history alias/supersession ve migration completion receipt.
- Schema contract, binary schema contract, schema compatibility generation, migration manifest readiness kontrolü veya custom minimum-version registry.
- Platform migrator runtime'ı, tenant migration executor servisi, migration worker'ı veya migration'a özel orchestration.
- Manifest üretici/doğrulayıcı, release hazırlayıcı, migration artifact kopyalayıcı veya bunlara ait CLI flag, Taskfile adımı, Docker katmanı ve CI job'u.
- Goose'un önüne yerleştirilmiş custom parser, planner, classifier, selector, verifier, finalizer veya migration state machine.
- Aynı DDL'yi hem Foundation hem Goose hem de uygulama kodunda yöneten paralel yollar.

Bir bileşen yalnızca eski bileşenin yerine yeni isimle taşınmışsa kaldırılmış sayılmaz. Eski kavramları gizleyen yeni abstraction da yasaktır.

## Veri koruma ve mevcut kurulumlar

Bu çalışma mevcut tenant'ları, platform/business tablolarını veya kullanıcı verisini silme/recreate etme çalışması değildir.

- Mevcut PostgreSQL database'i yeniden oluşturulmayacaktır.
- Var olan business tabloları, satırlar, tenant schema'ları, foreign key ilişkileri ve dosya referansları korunacaktır.
- Eski migration framework'üne ait metadata tabloları/function/procedure'lar ancak yeni Goose akışı doğrulandıktan ve hiçbir runtime consumer kalmadığı kanıtlandıktan sonra Goose forward migration'larıyla kaldırılır.
- Eski migration metadata'sı business data gibi korunmaz; tamamen emekliye ayrılır.
- Mevcut schema'yı Goose history'ye alma işlemi yalnızca Goose'un native mekanizmasıyla yapılır. Goose dışı bir baseline registry veya özel adoption tablosu kurulmaz.
- Uygulama rollback'i schema'yı eski migration framework'üne döndürmek anlamına gelmez. Gerekli düzeltmeler yeni forward Goose migration'larıyla yapılır.
- DROP DATABASE, tenant schema drop/recreate, business table truncate veya veriyi silip yeniden seed etme kabul edilmez.

## Cursor/implementasyon ajanı için çalışma emri

1. Önce repository'nin tamamını tara. Yalnızca tek bir migrator package'ına bakarak kapsamı daraltma.
2. Go source, SQL, contracts, JSON Schema, generated output, tests, fixtures, Dockerfile, Compose, Taskfile, scripts, CI, setup CLI, skills ve docs dahil tüm referansları takip et.
3. Eski sistemi korumak için KEEP/compatibility/fallback kararı verme. Eski migration framework'üne ait her kod ve veri modeli silinecek veya Goose native akışına dönüştürülecektir.
4. Kod değişikliğini yalnızca bu dört aşamaya göre yap. Yeni bir beşinci süreç/servis/adım gerekiyormuş gibi görünürse bunu oluşturma; sorumluluğu dört aşamadan doğru olana yerleştir.
5. İlk aşamada yalnızca envanter çıkarıp bırakma. Envanter, silme ve yeniden kurma işinin başlangıcıdır; tamamlanma kriteri değildir.
6. Başarısız testi atlama, guard'ı gevşetme, eski behavior'ı test adı altında saklama veya yalnızca derlemeyi başarı sayma.
7. Çelişen eski skill/architecture metinlerini güncelle. Kodun yeni mimariye taşındığını söyleyip eski dokümanlarda eski sistemi normatif bırakma.
8. Tüm kontrolleri gerçek PostgreSQL ile çalıştır. Mock veya statik test tek başına migration doğrulaması değildir.

## İş sırası

- [01 — INITIALIZE](01-initialize.md)
- [02 — DATABASE FOUNDATION](02-database-foundation.md)
- [03 — GOOSE](03-goose.md)
- [04 — APPLICATION ve eski sistemin tamamen kaldırılması](04-application-and-total-removal.md)

## Nihai kabul koşulu

İş ancak aşağıdaki koşulların **tamamı** sağlandığında biter:

- Runtime lifecycle tam olarak dört aşamadan oluşur.
- Goose, platform ve tenant schema değişikliklerinin tek otoritesidir.
- Goose native history dışında migration geçmişi yoktur.
- Eski migration framework'üne ait executable code, model, contract, SQL routine, table, script, build artifact, service, test fixture veya aktif doküman kalmamıştır.
- Temiz database kurulumu ve mevcut database/tenant upgrade'i gerçek PostgreSQL'de doğrulanmıştır.
- Mevcut veriler korunmuştur.
- task dev ilk kurulumda ve tekrar çalıştırmada başarılıdır.
- API, worker ve diğer uygulama servisleri Goose tamamlanmadan başlamaz.
- Kod ve dokümantasyon aynı mimariyi tarif eder.
- Test, fileguard, archguard ve CI kapıları geçer.

Bu koşullardan biri eksikse iş tamamlanmış değildir.
