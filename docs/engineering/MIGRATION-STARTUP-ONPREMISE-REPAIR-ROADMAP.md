# Harpy — Migration, Startup Diagnostics ve On-Premise Bootstrap Düzeltme Planı

> Amaç: migration manifest üretimini tek komuta bağlamak, startup hatalarını sebebiyle görünür kılmak, yalnız açılışı durduran SQL assert'lerini anlamlı hata veren kontroller yapmak ve single-tenant/on-premise kurulumun kendi geçerli installation manifest + lisans offering verisini üretmesini sağlamak.
>
> Sınır: Bütün eski `1/0` assert'lerini toplu yeniden yazma. Sadece API başlangıcını durduran assert'lerde ve doğrudan ilgili kontrollerde değişiklik yap.

## Kod üzerinden doğrulanan mevcut durum

- `core/internal/composition/api/start_api.go`: DB pool açma, deployment policy bağlama ve readiness başarısızlıklarının üçünde `reportFailure(nil, logger, "dependency_preflight")` çağrısı var; hata nesnesi/cause loglamaya aktarılmıyor.
- `core/internal/composition/api/report_failure.go`: `reportFailureCause` zaten `cause` alanını destekliyor. Startup dalları bunu kullanmalı; yeni logging altyapısı gerekmiyor.
- `core/internal/modules/billing/migrations/platform/fresh/20260722141100_create_billing_installation_offering_read_v1.fresh.sql`: `onpremise` dalı `billing_license_heads` ve `billing_license_offering_publications` eşleşmesinden tam bir kayıt bekliyor. Sonuç sayısı 1 değilse SQLSTATE `55000`, `billing installation offering is unavailable` hatası veriyor. SaaS dalı ayrı olarak `billing_saas_control_offering_publications` okuyor.
- `cmd/setup/runtime_authority.go`: `writeAuthorityMaterial` her iki profile için `installation.json` dosyasına `{}` yazıyor.
- `contracts/installation/v1`: `installation-manifest.v1` manifest'i environment ID, deployment mode, primary tenant, first admin ve request digest bekliyor. `{}` geçerli değil.
- `cmd/setup/main.go`: setup sonrası sıradaki komutlar migrator `deploy up`, operator `initialize`, ardından API. Operator initialize hatası bu akışı kesiyor.
- Fresh migration manifest `contracts/runtime/v1/fresh-migration-manifest.json`; binary contract içindeki `migration_manifest_sha256` alanı onun digest'i ile eşleşmeli.

Not: Plan mevcut branch'te görülen dosyalara ve tarif ettiğin arıza zincirine dayanır. `checkfresh` implementasyonunun gerçek yolunu repo ağacında ayrıca bulup düzenle; mevcut `checkfresh -write` davranışını tarifinden biliyoruz.

## Faz 0 — Güvenli başlangıç ve arızayı tekrarlama

1. Değişiklikleri ayrı bir branch üzerinde yap.
2. Başlangıç durumunu ve testleri kaydet:
```bash
git status --short
go test ./contracts/installation/v1 ./cmd/setup ./core/internal/composition/api
```
3. Mevcut komutları ve başarısızlıkları kaydet. İlk aşamada SQL migration'larını canlı DB'de çalıştırma.
4. Repo'nun gerçek CI/task ve migration bundle doğrulama komutlarını bul; bunları referans test olarak kullan.

**Çıkış kriteri:** Çalışma ağacı ve başlangıç test sonuçları biliniyor; eski hatalar yeni değişikliklerden ayrıştırılabiliyor.

## Faz 1 — Fresh migration manifest'ini tek komutla güncelle

### İstenen CLI davranışı

Yeni migration eklendiğinde geliştirici tek komut çalıştırmalı:
```bash
go run ./internal/migrationbundle/cmd/checkfresh -write
```
Komut gerçek migration dosyalarından digest'i hesaplamalı, `binary-schema-contract.json` içindeki `migration_manifest_sha256` alanını güncellemeli ve `fresh-migration-manifest.json` dosyasını yeniden üretmeli. Contract'taki eski hash'in yeni hash ile önceden eşit olması bir önkoşul olmamalı; bu, güncelleme aracını döngüsel olarak kilitler.

### Uygulama adımları

1. `internal/migrationbundle/cmd/checkfresh` implementasyonunu ve manifest builder'ını bul.
2. “Manifest'i hesapla → mevcut contract hash'i ile karşılaştır → farklıysa yazmayı reddet” sırasını kaldır.
3. Tek canonical builder kullan:
   - Mevcut keşif/filtreleme kurallarına göre migration dosyalarını topla.
   - Her SQL dosyasının SHA-256'sını gerçek byte içeriğinden hesapla.
   - Manifest'i kararlı biçimde serialize et ve nihai manifest bytes üzerinden manifest digest'ini hesapla.
   - Manifest dosyasıyla contract'taki `migration_manifest_sha256` alanını birlikte güncelle.
4. İki dosyalı yazımı mümkün olduğunca atomik yap: geçici dosyalara yaz, parse/hash doğrula, sonra rename et. Hata durumunda yarım yazılmış state bırakma; gerekirse rollback uygula.
5. `-write` dışındaki check modu salt-okunur kalsın; drift varsa hangi dosya/alanın neden uyuşmadığını yazıp non-zero dönsün.
6. Supersession kayıtlarını gerçek dosyalarla doğrula:
   - `old_sha256` eski SQL dosyasının gerçek SHA-256'sı olmalı.
   - `replacement_sha256` replacement SQL dosyasının gerçek SHA-256'sı olmalı.
   - Dosya yoksa veya referans çözülemiyorsa hash uydurma; path'i belirterek fail et.
7. Regression test: contract hash'ini eski değere ayarla, SQL/manifest içeriğini değiştir, `-write` iki dosyayı güncellesin.
8. Supersession fixture'ında eski/yeni SQL byte'larından gerçek digest hesaplandığını test et.
9. Idempotency testi: değişiklik yoksa ikinci çalıştırma gereksiz diff oluşturmamalı.

**Kabul kriterleri:** Hash elle kopyalanmıyor; eski contract hash'i yüzünden `-write` başarısız olmuyor; manifest ile contract aynı anda doğru değere geliyor; supersession hash'leri gerçek dosyalardan hesaplanıyor; check modu dosya değiştirmiyor.

## Faz 2 — API startup hatalarının sebebini güvenli logla

### Değiştirilecek dallar

`core/internal/composition/api/start_api.go` içinde:
1. `openPool(...)` hata döndürmesi.
2. `bindDeploymentPolicy(...)` hata döndürmesi.
3. `probeReadiness(...)` ardından `allReady(initial)` false olması.
Ayrıca açılışı durduran deployment mode parse ve `health.NewMonitor` dallarının da güvenli hata nedeni yazması sağlanmalı.

### Uygulama adımları

1. İlk iki dalda `reportFailureCause` kullan veya aynı sözleşmede cause taşıyan bir helper ekle. Mevcut `report_failure.go` bunu zaten destekliyor.
2. Readiness başarısızlığında `initial` içinden hangi kontrollerin başarısız olduğunu çıkar. Mevcut readiness tiplerine göre `failed_check`, `reason_code`, `cause` gibi alanlar kullan; ikinci bir readiness sistemi kurma.
3. Ham `err.Error()` değerini doğrudan loglama. DSN/URL user-password, secret, token, key, Authorization header ve query parametre değerlerini çıkar. Bilinen hata tiplerinden allowlist edilmiş güvenli alanlar logla: aşama, bağımlılık türü, SQLSTATE, timeout/canceled, readiness check adı. Bilinmeyen hatada ham metin yerine hata sınıfı ve güvenli aşama kodu kullan.
4. `error_class=dependency_preflight` alanını koru; bu kararlı sınıflandırmadır. İnsan tarafından okunabilir nedeni ayrı alan olarak ver.
5. Logger çıktısını yakalayan testler ekle: DB pool, policy bind ve readiness başarısızlıklarını ayrı ayrı test et. Sahte DSN parolası/token çıktıda bulunmamalı; aşama ve güvenli sebep bulunmalı.

**Kabul kriterleri:** Log sadece `dependency_preflight` demiyor; hangi aşamanın neden başarısız olduğu anlaşılabiliyor; sırlar görünmüyor; exit code ve hata sınıflandırması korunuyor.

## Faz 3 — Yalnız açılışı durduran assert'leri anlaşılır yap

### Kapsam

Bütün repo genelinde `SELECT 1/0` veya eşdeğerlerini topluca değiştirme. API açılışını doğrudan engelleyen platform/billing/operator initialize zincirindeki assert ve preflight sorgularını belirle.

### Uygulama adımları

1. Startup sırasında çalıştırılan migration/assert/readiness sorgularını çağrı zincirinden çıkar.
2. Her assert'in neyi kontrol ettiğini kaydet: tablo, rutin, rutin imzası, beklenen satır sayısı veya contract version.
3. Bölme hatası yerine açıklayıcı exception üret. PL/pgSQL kullanılıyorsa örnek:
```sql
RAISE EXCEPTION USING
  ERRCODE = '55000',
  MESSAGE = 'startup assertion failed: expected exactly one offering record',
  DETAIL = 'expected=1 actual=' || v_actual::text;
```
Mesajı ilgili kontrolün gerçeğine göre yaz; düz SQL statement içinde uygun olmayan PL/pgSQL sözdizimi kullanma.
4. Eksik tablo için schema/table, eksik routine için tam isim + identity arguments, sayım assert'i için expected/actual belirt.
5. Yayımlanmış migration'ların checksum'ını bozacak şekilde yerinde değiştirme. Gerekliyse ileri migration ekle; fresh-install için gereken değişiklik manifest kurallarıyla birlikte gelsin.
6. Yalnız ilgili kontroller için negatif test ekle: nesne yok, imza yanlış, sayım 0/2.
7. Hata detail'inde secret veya bağlantı bilgisi olmadığını denetle.

**Kabul kriterleri:** Açılışı durduran assert hatasında isim/imza veya expected/actual görülebiliyor; toplu refactor yapılmıyor; uygulanmış migration checksum'ları bozulmuyor.

## Faz 4 — Setup geçerli installation manifest'i üretsin

### Gerçek problem

`cmd/setup/runtime_authority.go` içindeki `writeAuthorityMaterial` `installation.json` dosyasına `{}` yazıyor. Oysa `contracts/installation/v1` geçerli `installation-manifest.v1` için environment ID, deployment mode, primary tenant, first admin ve request digest bekliyor. Operator'ın sonradan elle JSON doldurmasını isteme; setup metadata'yı toplamalı/üretmeli ve sözleşme paketini kullanmalı.

### Uygulama adımları

1. `cmd/setup/answers.go`, prompt ve profile yazma akışına şu metadata'yı ekle/bağla:
   - Environment UUID: güvenli üret, her kurulumda sabit kalsın ve tekrar setup çalışınca korunsun.
   - Deployment mode: `single-tenant` profile'ını runtime'ın beklediği `onpremise` değerine açıkça map et. `single` tenancy modu ile `onpremise` deployment mode'u aynı kavram değil.
   - Primary tenant name/slug.
   - First admin display name/email.
2. Metadata setup sırasında kullanıcı girdisi veya açıkça tanımlanmış güvenli default'lardan gelsin. Gerekli değerleri sessizce uydurma; `--defaults` politikasıyla uyumlu şekilde ya açıkça tanımlanmış ilk kurulum varsayılanı kullan ya da anlaşılır hata üret.
3. `installationv1.Build(installationv1.BuildInput{...})` ile manifest oluştur. Ham `{}` yazma.
4. `Manifest.CanonicalBytes()` çıktısını mevcut durable/atomic file writer ile `secrets/runtime/installation.json` yoluna yaz.
5. Setup tekrarlandığında geçerli manifest'in EnvironmentID/kurulum kimliğini koru; secret rotation installation kimliğini değiştirmesin.
6. Manifest mode'u runtime config mode'u ile aynı olsun; config parse/validation testleri ekle.
7. Setup testleri: çıktı `installationv1.Parse` ile parse edilebilmeli, mode `onpremise` olmalı, zorunlu alanlar valid olmalı, tekrar setup kimliği korumalı ve `{}` üretilmediği doğrulanmalı.

**Kabul kriterleri:** Setup sonrası installation manifest manuel müdahale olmadan valid; kimlik tekrar setup'ta korunuyor; runtime ve manifest mode'u eşleşiyor.

## Faz 5 — On-premise billing lisans offering kaydını doğru yoldan provision et

### Semantik karar

`platform.billing_installation_offering_read_v1(..., 'onpremise')` SaaS offering tablosunu fallback olarak okumamalı. Bu deployment mode ayrımını bozar. On-premise kurulum geçerli yerel lisans/entitlement authority kaydı oluşturmalı; read function mevcut sözleşmesine göre o kaydı okumalı.

### Uygulama adımları

1. Billing migration'larında desteklenen bootstrap/publication akışını incele: `billing_license_heads`, `billing_license_offering_publications`, lisans snapshot ve offering publication rutinleri. Guard trigger ve yetki sınırlarını atlayarak tablolara doğrudan insert yapma.
2. Gerçek lisans kaynağını kod/contract üzerinden belirle:
   - İmzalı/offline license artifact var mı?
   - İlk kurulum için local/community entitlement veya setup lisansı üreten mevcut prosedür var mı?
   - Authority/issuer kimliği ve verification trust bundle nereden geliyor?
   Kaynağı bulmadan sahte `license_id`, hash, signature veya offering id üretme.
3. `cmd/setup` gerekli trust/key/artifact dosyalarını hazırlasın. Lisans yayınlama migration/initialize aşamasındaysa aynı desteklenen authority rutinlerini çağır.
4. Provisioning'i idempotent operator/bootstrap adımı yap:
   - Geçerli installation manifest'i yükle.
   - EnvironmentID ve `onpremise` mode'u doğrula.
   - Lisans/authority kaynağını doğrula ve offering'i yayınla.
   - Head ile publication arasındaki environment, license, snapshot sequence/hash ilişkisini doğrula.
   - Sonuç sayısının tam 1 olduğunu doğrula.
5. Bu adımı operator initialize'ın readiness/tenant assignment adımından önce çalıştır; migrator'ın yetki modelini aşma. Gerekirse `cmd/setup` `nextCommands` sırasını ve dokümantasyonu güncelle.
6. İkinci initialize çalıştırıldığında aynı kaydı yeniden üretmemesi/çakışmaması için idempotency testi ekle.
7. PostgreSQL integration senaryoları:
   - Geçerli on-prem license + manifest: offering read tek satır döndürür.
   - Head/publication yok: SQLSTATE `55000` ve güvenli teşhis.
   - Yalnız SaaS offering var: on-premise başarısız kalır, SaaS fallback yapılmaz.
   - Environment veya snapshot hash ilişkisi yanlış: kayıt reddedilir.

### Güvenlik noktası

Setup'ın imzasız entitlement veya herkese açık lisans uydurması çözüm değildir. Ürünün gerçek on-premise lisanslama modeli kullanılmalı. Model eksikse önce issuance/verification sözleşmesini tamamla; tabloya geçici veri basarak API'yi yeşile çevirme.

**Kabul kriterleri:** Fresh single-tenant install sonrasında offering read tam 1 satır döndürür; SaaS offering olmadan on-premise çalışır; yalnız SaaS offering varken on-premise başarısızdır; readiness ancak gerçek önkoşullar sağlanınca yeşil olur.

## Faz 6 — Uçtan uca regresyon ve CI

Aşağıdaki testler otomatik olmalı:
1. **Fresh database / single-tenant:** boş DB → migration deploy → setup manifest → operator initialize → API start/readiness green.
2. **Migration değişikliği:** test migration ekle → tek komut manifest + contract hash üret → check modu başarılı.
3. **Supersession integrity:** SQL'de bir byte değişince check başarısız; write sonrası tüm digest'ler gerçek dosyalarla eşleşir.
4. **Eksik billing offering:** on-premise head/publication yok → SQLSTATE 55000 ve anlaşılır teşhis.
5. **SaaS isolation:** SaaS offering mevcut ama on-premise lisansı yok → on-premise bootstrap lisans üretmedikçe fail.
6. **Log redaction:** sahte DSN parolası ve token startup loglarında yer almıyor.
7. **Idempotency:** setup ve operator initialize ikinci kez çalışınca installation ID veya license publication çakışmıyor.

Önerilen son doğrulama sırası:
```bash
go test ./contracts/installation/v1 ./cmd/setup ./core/internal/composition/api
# migrationbundle checkfresh testleri ve manifest check
# ilgili billing/operator integration testleri (PostgreSQL 18 fixture)
# fresh single-tenant end-to-end bootstrap
git diff --check
git status --short
```
Repo'nun gerçek CI komutlarını bununla birlikte çalıştır; yukarıdaki komutlar tek başına E2E doğrulamasının yerine geçmez.

## Uygulama sırası

| Sıra | İş | Neden |
|---|---|---|
| 1 | Startup cause logging | Teşhisi hızlı iyileştirir, riski düşük |
| 2 | Manifest tek-komut üretimi | Migration hash darboğazını kaldırır |
| 3 | Installation manifest üretimi | Operator initialize'a geçerli kurulum kimliği verir |
| 4 | Billing on-premise lisans/offering bootstrap | API'nin gerçek işlevsel blokajını çözer |
| 5 | Açılışı durduran assert'ler | Hataları nesne/imza/sayı seviyesinde açıklar |
| 6 | Fresh install E2E + CI | Dört konuyu tek regresyon senaryosuyla korur |

## Yerel agent için kesin talimatlar

- Her fazı ayrı commit olarak tamamla.
- İlk işte geniş refactor yapma; ilgili dosya ve çağrı zincirini oku.
- Daha önce yayımlanmış migration dosyalarını checksum'ı bozacak şekilde değiştirme.
- SHA-256, signature, UUID, lisans/offering ve manifest değerlerini uydurma; gerçek kaynaktan üret.
- Test başarısızsa testi gevşetme; sebebi bulup uygulamayı düzelt.
- Kod değişikliğini tamamlanma sayma. Her fazın kabul kriterini test çıktısıyla kanıtla.
- Son raporda değişen dosyalar, çalıştırılan komutlar, sonuçlar, kalan engeller ve migration'ların fresh/upgrade etkisini yaz.
