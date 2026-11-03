# 01 — INITIALIZE

## 1. Amaç

INITIALIZE, yeni Harpy database lifecycle'ının tek giriş noktasıdır. Bu giriş noktası mevcut initialize/orchestrator sisteminin refactor edilmiş sürümü olmayacaktır. Eski initialization ve migration orchestration kodu kaldırılacak; yeni giriş noktası dört aşamayı sıfırdan çağıracak şekilde yazılacaktır.

Tek ve değiştirilemez sıra:

1. INITIALIZE
2. DATABASE FOUNDATION
3. GOOSE
4. APPLICATION

INITIALIZE aşama sıralamasını yönetir. Schema yönetmez, SQL migration keşfetmez, migration seçmez, history tutmaz ve kendi completion state'ini üretmez.

## 2. Giriş noktası kontratı

Repository'de tek bir açık kurulum komutu/entrypoint bulunmalıdır. Local development, container startup ve otomatik kurulum aynı entrypoint'i kullanır. Aynı iş için birden fazla bootstrap veya migrate komutu bulunamaz.

Giriş noktası aşağıdaki sırayı uygular:

1. Konfigürasyonu ve gerekli bağlantı parametrelerini doğrula.
2. Hedef database'in yeni kurulum için uygun, boş bir database olduğunu doğrula.
3. DATABASE FOUNDATION işlemini çağır.
4. Foundation başarılıysa GOOSE işlemini çağır.
5. Goose başarılıysa APPLICATION süreçlerini başlat.
6. Her aşamanın sonucunu açık biçimde raporla ve gerçek hata kodunu dışarı aktar.

Herhangi bir aşama başarısız olduğunda sonraki aşamaya geçilmez. Foundation veya Goose hata verdiyse API/worker process'i başlatılmaz. Hata, generic bir initialize failed mesajına indirgenerek kök sebep kaybettirilmez.

## 3. Sıfırdan kurulum davranışı

INITIALIZE yeni ve boş database için tasarlanır. Eski sistemin database'ini veya schema'sını algılayıp yeni sisteme uyarlamaya çalışmaz.

Aşağıdaki durumlarda fail closed davranışı zorunludur:

- Hedef database erişilemiyorsa.
- Database yeni kurulum için boş değilse.
- Beklenmeyen uygulama veya migration framework tabloları varsa.
- Hedef schema veya role yapısı beklenen temiz kurulum durumuyla uyuşmuyorsa.
- Gerekli credential bulunamıyorsa.
- Foundation veya Goose başarısızsa.

Bu koşullarda INITIALIZE eski sistemi tanımaya, eski state'i dönüştürmeye, Goose version numarasını ileri almaya, özel baseline oluşturmaya veya mevcut database'i otomatik temizlemeye çalışmaz.

**Boş olmayan database'i otomatik olarak drop etme.** Yanlış hedefe bağlanıldığında veri silinmesini önlemek için işlem durmalı ve açık hata vermelidir. Fresh kurulum, eski database'in üzerine yazmak değil, açıkça ayrılmış yeni ve boş bir database kullanmaktır.

## 4. Idempotency ve tekrar çalıştırma

INITIALIZE aynı yeni kurulum üzerinde tekrar çalıştırılabilir olmalıdır. Tekrar çalıştırma için özel state registry oluşturulmaz.

Idempotency şu gerçek kaynaklar üzerinden sağlanır:

- PostgreSQL database/schema/role/extension nesnelerinin mevcut durumu.
- Foundation işlemlerinin güvenli create-if-not-exists ve doğrulama davranışı.
- Goose native version history.
- Application process manager'ın mevcut servis durumu.

Bir aşama yarıda kesildiyse tekrar çalıştırma yalnızca eksik Foundation nesnelerini tamamlar ve Goose'u yeniden çağırır. Uygulanmış Goose migration'ı yeniden uygulanmaz. Var olan nesneleri silerek temiz kurulum taklidi yapılmaz.

INITIALIZE aşağıdakilerden hiçbirini oluşturamaz:

- Initialization state tablosu
- Initialize run/attempt/receipt/checkpoint kaydı
- Database veya runtime generation değeri
- Release veya migration manifest kaydı
- Custom completed-stage registry
- Goose dışı migration version kaydı

## 5. Eski akışın kaldırılması

Mevcut initialize giriş noktasını satır satır düzeltme veya eski orchestrator'ı yeni bir wrapper arkasında tutma. Eski kurulum ve migration akışını çağıran bütün kodu sil ve gerekli yeni kodu temiz bir implementation olarak oluştur.

Kontrol edilecek alanlar:

- Go command/main package'ları
- initialize/bootstrap/operator/orchestrator package'ları
- Eski migration servislerine yapılan process ve RPC çağrıları
- CLI flag ve subcommand'ları
- Taskfile task'ları
- Docker entrypoint ve CMD/ENTRYPOINT tanımları
- Compose dependency ve healthcheck sıralaması
- Local development script'leri
- CI setup ve integration test adımları
- Environment variable ve config struct'ları
- Eski artifact, manifest veya release dosyası path'leri

Aynı işlemi yapan ikinci bir yol kalırsa kaldır. Bir eski fonksiyonu yeni entrypoint'ten çağırmak, eski sistemi kaldırmak değildir.

## 6. Hata modeli

Hatalar hangi aşamada meydana geldiyse o aşamaya ait, korunmuş sebep ve bağlamla raporlanmalıdır:

- Konfigürasyon/bağlantı doğrulama hatası
- DATABASE FOUNDATION hatası
- GOOSE migration hatası
- APPLICATION process başlangıç hatası

Hata raporu en azından aşama, hedef database/schema, işlemin güvenli biçimde tekrar denenebilir olup olmadığı ve alttaki hata bilgisini içermelidir. Password, DSN secret veya token loglanamaz.

Hata sınıflandırması için migration-specific custom state machine oluşturulmaz. Goose hatası Goose çağrısının gerçek hata sonucu olarak korunur. Foundation hatası ilgili PostgreSQL işleminin hatası olarak korunur.

Bir Goose migration başarısızsa Application başlamaz. Goose'un native history davranışı dışındaki bir checkpoint/reconcile tablosu ile devam kararı verilmez.

## 7. APPLICATION başlatma sınırı

APPLICATION yalnız Goose başarıyla döndükten sonra başlatılır. API veya worker kendi startup akışında Foundation ya da Goose çağırmaz.

Normal API/worker restart'ı yalnız uygulama process'ini başlatır. Restart sırasında:

- Goose komutu çalıştırılmaz.
- Schema kontrolü adı altında DDL yürütülmez.
- Eksik tablo otomatik oluşturulmaz.
- Foundation tekrar tetiklenmez.
- Eski migration state'i okunmaz.

## 8. Konfigürasyon

INITIALIZE için yalnızca gerçekten ihtiyaç duyulan bağlantı ve çalışma ayarları bulunur:

- Yeni, boş hedef PostgreSQL database'i için bağlantı bilgisi.
- Foundation işlemleri için gerekli ayrıcalıklı bağlantı bilgisi.
- Goose'un hedef database/schema ve native migration source bilgisi.
- Application servislerinin çalıştırılması için gereken process/config bilgisi.

Eski manifest path, archive path, digest, release ID, generation, catalog veya migration run alanlarının hiçbir karşılığı yeni config'e eklenmez. Eski env isimlerini alias olarak kabul eden uyumluluk kodu yazılmaz.

## 9. Zorunlu testler

Gerçek PostgreSQL kullanarak:

1. Boş database'de INITIALIZE sırasının tam dört aşama olduğunu doğrula.
2. Foundation başarısız olduğunda Goose ve Application'ın çalışmadığını doğrula.
3. Goose başarısız olduğunda Application'ın çalışmadığını doğrula.
4. Tam kurulumdan sonra INITIALIZE tekrar çağrıldığında duplicate DDL/data oluşmadığını doğrula.
5. Foundation ortasında kesilen kurulumun tekrar çağrıldığında güvenle tamamlandığını doğrula.
6. Goose ortasında hata alan kurulumun Goose native history ile tekrar denenebildiğini doğrula.
7. Boş olmayan/eski bir database hedeflendiğinde işlemin durduğunu doğrula.
8. API/worker process restart'ının Goose çağırmadığını doğrula.
9. Hata mesajında alttaki PostgreSQL/Goose hata sebebinin kaybolmadığını doğrula.
10. Aynı işi yapan ikinci bir initialize/migrate entrypoint olmadığını doğrula.

## 10. Tamamlanma şartı

Tek bir yeni INITIALIZE entrypoint bulunur. Bu entrypoint yalnızca dört aşamayı sırasıyla çağırır. Eski initialization/migration orchestration kodu ve onun tüm tüketicileri kaldırılmıştır. Boş olmayan/eski database otomatik dönüştürülmez. Tekrar çalıştırma özel state tablosuna değil gerçek PostgreSQL nesnelerine ve Goose native history'ye dayanır.

## 11. Repository'deki mevcut INITIALIZE consumer envanteri

Uygulama öncesi aşağıdaki gerçek çağrı zincirini yeniden çıkar; eski task adları varsayılarak değil YAML task graph ve container entrypoint üzerinden doğrulanarak çalışılmalıdır.

| Alan | Güncel dosya / giriş | Yapılacak değişiklik |
| --- | --- | --- |
| Root entrypoint | `Taskfile.yml` | `task dev` yalnız `:local:up` üzerinden tek initialize akışına gider. |
| Local lifecycle | `.taskfiles/local.yml` | `bootstrap` ve `migrate` duplicate zincirini kaldır; tek `initialize` çağrısı kur. |
| Recovery rehydrate | `.taskfiles/local-recovery.yml` | `:local:bootstrap` + `:local:migrate` yerine initialize entrypoint'ini çağır; restore öncesi DB boşluğu ve restore sırasını gerçek PostgreSQL ile kanıtla. |
| Local operations | `.taskfiles/local-operations.yml` | Migrator/schema/tenant executor log/status/check bağımlılıklarını kaldır. |
| API / worker | `ops/local/compose.runtime.yml` | API ve worker yalnız initialize tamamlandıktan sonra başlar; schema-executor dependency'si kaldırılır. |
| Compose tools | `ops/local/compose.tools.yml` | Tek initialize one-shot service; migrator ve tenant executor probe'ları yok. |
| Local/production image | `ops/local/Dockerfile`, `ops/production/Dockerfile` | Eski migrator binary ve manifest artifact'i paketlenmez. |

### 11.1 Tek girişin kesin çalışma kontratı

Yeni entrypoint:
- Bir kez configuration/preflight yapar.
- Hedef DB kimliğini ve boşluk koşulunu doğrular.
- Foundation'ı çağırır.
- Aynı Goose implementation ile platform source'unu, sonra gerekiyorsa tenant source'unu çağırır.
- Yalnız başarıdan sonra APPLICATION servislerini başlatır.
- Hata durumunda aşama, güvenli retry niteliği ve redacted underlying cause döndürür.
- Kendine ait migration/init run, generation, checkpoint, receipt veya completion tablosu yazmaz.

Installation artifact, license, initial credential delivery ve recovery signing gibi mevcut installation domain davranışları sırf migration koduyla aynı binary'de diye silinmez. Bunların owner'ı ayrı tutulur ve INITIALIZE sırasındaki bağımlılıkları açıkça tanımlanır. Bu işlemlerin hiçbiri Goose version/history veya migrator release ledger'ı tüketemez.

### 11.2 Legacy consumer kapatma prosedürü

Her eski entrypoint için:
1. Çağıran task/service/command bulunur.
2. Tüketilen config/secret/env alanları bulunur.
3. Generated artifact producer'ı bulunur.
4. Healthcheck, depends_on, volume ve port referansları bulunur.
5. Test/fixture/contract ve docs referansları bulunur.
6. Yeni entrypoint veya kaldırma kararı consumer seviyesinde uygulanır.
7. Arama tekrar çalıştırılır; geride kalan her eşleşme disposition alır.

### 11.3 Tamamlanma

INITIALIZE tamamlandı kabul edilebilmesi için eski `core-migrator` entrypoint'ine, `local-release prepare`, migration manifest, `deploy up/verify`, `bootstrap`, `migrate`, tenant/schema executor, custom finalization ve recovery rehydrate eski zincirine referans kalmamış olmalıdır. Backup/restore komutları ayrı migration dışı sahipliğe taşınmış ve recovery testleri geçmiş olmalıdır.
