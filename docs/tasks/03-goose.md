# 03 — GOOSE

## Amaç

Goose'u Harpy'nin bütün schema değişiklikleri için tek ve doğrudan migration otoritesi yap. Platform schema'ları ve schema-per-tenant yapısındaki bütün tenant schema'ları aynı Goose mekanizmasıyla yönetilir.

Bu görevde custom migration framework'ü sadeleştirerek yaşatmak yoktur. Custom framework tamamen kaldırılır; yalnız Goose'un kendi migration discovery, execution ve native version history davranışı kalır.

## Tek migration otoritesi

Goose aşağıdaki tüm PostgreSQL DDL değişikliklerini yönetir:

- CREATE / ALTER / DROP TABLE
- Column ve data type değişiklikleri
- Index, unique constraint, check constraint ve foreign key
- Sequence, enum ve composite type
- View, function, procedure ve trigger
- Platform schema değişiklikleri
- Tenant schema değişiklikleri
- Foundation sonrası kurulacak uygulama tabloları
- Eski migration framework metadata'sının sonradan kaldırılması

Bu listede bulunmayan ikinci bir DDL yönetim yolu oluşturma.

## Native Goose kuralları

- Resmi github.com/pressly/goose/v3 API/CLI kullan.
- Migration dosyaları Goose'un native SQL migration formatında olmalıdır.
- Migration discovery normal filesystem/embed fs.FS üzerinden yapılmalıdır.
- Versiyon ve sıralama Goose tarafından belirlenmelidir.
- Applied migration dosyaları değiştirilmez; hata yeni bir forward migration ile düzeltilir.
- Goose native PostgreSQL version store tek migration history'dir.
- Platform schema ve her tenant schema için Goose'un native history mekanizması kullanılır.
- Tenant schema history'si yalnız Goose tarafından yazılır/okunur.
- Goose native API'nin sağladığı seçenekler dışında custom store/ledger geliştirilmez.
- Goose invocation mümkün olduğunca tek, küçük, test edilebilir bir adapter/command sınırında tutulur. Bu adapter yalnız Goose'u doğru DSN/schema/source ile çağırır; kendi migration policy veya history modeli bulunmaz.
- SQL migration'lar transaction davranışı ve lock gereksinimleri açıkça tanımlanarak yazılır.
- Goose'un native locking kabiliyeti yeterli değilse yalnız concurrent invocation'ı engelleyen PostgreSQL advisory lock kullanılır. Lock için custom state tablosu oluşturulmaz.

## Platform ve tenant migration source'ları

Repository'deki mevcut SQL dosyalarını tarayıp aşağıdaki hedefe dönüştür:

- Platform schema'ları için tek native Goose migration source.
- Tenant schema'ları için tek native Goose migration source.
- Her iki source aynı Goose implementation ve native history davranışını kullanır.
- Platform ve tenant SQL'leri ownership/scope bakımından ayrılabilir; bu ayrım ikinci bir migration engine anlamına gelmez.
- Migration dosyaları manifest, owner catalog, path allowlist veya release metadata tarafından seçilmez.
- Aynı DDL'nin birden fazla aktif kopyası bulunmaz.
- Superseded/duplicate SQL corpus'u yeni sisteme taşınmaz.
- Gerekli schema oluşturma ve tüm DDL migration kaynaklarında açıkça bulunur.

Mevcut .fresh.sql corpus'unu otomatik olarak koruma. Her dosyayı gerçek DDL etkisi açısından sınıflandır. Goose'un normal native migration discovery mekanizmasıyla çalışabilecek, gerçekten gerekli ve tekil SQL migration'ları tut. Gereksiz kopyaları, superseded dosyaları, custom selector'ın elemek zorunda kaldığı duplicate'leri ve yalnız eski framework'ün okuyabildiği formatları kaldır. Gerekli mevcut schema'yı tek temiz Goose başlangıç migration'ı ve sonrasındaki gerçek forward migration'lar olarak yeniden kurmak uygunsa bunu yap; bu işlem mevcut database verisini silme gerekçesi değildir.

## Platform migration akışı

1. INITIALIZE, Foundation'ın başarılı olduğunu doğrular.
2. Goose platform migration source'unu native şekilde açar.
3. Platform schema hedefi için Goose native Up çalıştırılır.
4. Başarı/başarısızlık Goose native sonucu üzerinden raporlanır.
5. Başarısızlıkta Application başlatılmaz.
6. Mevcut platform Goose history dışındaki özel doğrulama/history kaydı tutulmaz.

## Tenant migration akışı

Her tenant schema aynı Goose runner ile hedeflenir:

1. Canonical TenantID ve kayıtlı schema name doğrulanır.
2. Foundation yalnız schema namespace'i ve ownership/privilege altyapısını hazırlar.
3. Goose tenant migration source'unu native şekilde açar.
4. Goose tenant schema için Up çalıştırır.
5. Goose başarılı olursa tenant uygulamaya açılır.
6. Goose başarısız olursa tenant aktif hale gelmez; gerçek hata loglanır ve retry Goose native history üzerinden yapılır.

Tenant migration için ayrı executor process, daemon, queue consumer, manifest, run/attempt/checkpoint kaydı veya özel completion SQL function oluşturma. Tenant oluşturma işinin başka bir business lifecycle ihtiyacı gerçekten varsa yalnız o genel iş amacına ait alanlar korunabilir; schema version/history/migration attempt alanları kaldırılamaz. Goose çağrısı her durumda doğrudan aynı Goose runner'dan yapılır.

## Eski framework'ün Goose'a taşınması

Önce tüm aktif migration SQL'lerini, mevcut platform/tenant schema'larını ve Goose native history durumunu gerçek PostgreSQL üzerinde karşılaştır.

Yeni/temiz kurulum için:
- Foundation altyapıyı kurar.
- İlk Goose migration seti tüm gerekli platform ve tenant tablolarını oluşturur.
- Goose native history hangi migration'ların uygulandığını kaydeder.
- Application bundan sonra başlatılır.

Mevcut kurulum için:
- Database ve tenant schema'ları korunur.
- Mevcut schema'nın hedef Goose baseline ile eşleştiği gerçek DB catalog üzerinden doğrulanır.
- Goose native version store, yalnız Goose'un resmi API/CLI mekanizmasıyla mevcut schema seviyesine alınır. Custom adoption table, schema contract veya manifest eklenmez.
- Ardından yalnız eksik forward Goose migration'ları çalıştırılır.
- Eski migration metadata'sı için gereken DROP işlemleri Goose forward migration'ında yapılır.
- Business tablolarına ve business satırlarına destructive işlem uygulanmaz.

Schema kontrolü için geçici/integration testleri ve PostgreSQL catalog sorguları kullanılabilir. Fakat bu kontroller runtime'da ikinci bir permanent schema-contract/version authority olarak tutulamaz.

## Tamamen kaldırılacak Goose dışı katmanlar

Aşağıdaki davranışlardan herhangi biri kaldıysa görev tamamlanmamıştır:

- Goose öncesi migration manifest okuma veya hash doğrulama
- SelectedMigration, MigrationManifest, custom migration type/catalog
- Migration dosyalarını seçen owner/scope/path/origin/supersession algoritmaları
- History alias, custom history classification, migration supersession
- Release manifest/archive/registry veya migration release resolver
- Migration run lifecycle, attempt, receipt, checkpoint, requirement
- Goose dışı migration status veya current-version table
- Custom Goose database store veya custom version store
- Schema contract/binary schema contract ve migration digest readiness
- Tenant-specific migrator/executor/service/worker/CLI
- Goose migration'ını tamamlayan custom DB function veya completion routine
- Migration hash, release ID veya generation bilgisini tenant claim/job/receipt'e taşıma
- Docker image'a migration manifest copy etme veya manifest path argümanı
- Migration manifest üreten/verifying tool, CI check veya Taskfile komutu

Bir eski struct'ın yalnız bir alanını veya bir SQL kolonunu kaldırmak yetmez. O alanın üreticisi, consumer'ı, JSON/SQL kontratı, test fixture'ı, generated output'u ve deployment wiring'i de silinmelidir.

## Test ve kabul

Gerçek PostgreSQL integration testleri zorunludur:

1. Boş database: Foundation ardından platform Goose migration'ları tüm gerekli schema'yı oluşturur.
2. Tenant A ve Tenant B: ikisine de aynı tenant Goose kaynağı uygulanır ve ayrı native history oluşur.
3. Aynı migration ikinci kez: Goose no-op davranışı; tekrar DDL veya data duplication yok.
4. Yarıda hata: Goose native history ile güvenli tekrar; özel checkpoint yok.
5. Platform ve tenant migration'larının sıralaması native Goose tarafından yönetilir.
6. Applied migration değiştirilirse test/guard bunu yakalar veya yeni forward migration kuralı uygulanır.
7. Mevcut DB: business table ve satırlar korunarak Goose native baseline'a alınır.
8. Eski metadata temizliği sonrası eski runtime nesnelerine hiçbir consumer kalmamıştır.
9. Runtime role DDL yapamaz; Goose rolü gerekli schema'larda migration yapabilir.
10. Goose tamamlanmadan Application readiness başarılı olamaz.

## Tamamlanma şartı

Goose dışında migration discovery, planlama, seçme, uygulama, version history, checkpoint, release, verification veya completion logic'i bulunmaz. Platform ve tenant schema'ları yalnız Goose ile değişir. Goose dışı migration framework'üne ait hiçbir kod, DB state modeli veya servis kalmaz.
