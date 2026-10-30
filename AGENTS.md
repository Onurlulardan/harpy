# Harpy

Harpy; single-tenant/on-premise ve multi-tenant/SaaS deployment modellerini destekleyen, Go ile geliştirilmiş modüler bir backend platformudur. Business yetenekleri tek bir Go codebase içindeki modüller olarak düzenlenir; API, worker, migrator ve operator bu modüllerden composition root'ları aracılığıyla oluşturulan ayrı runtime süreçleridir. Standart deployment altyapısı PostgreSQL 18, RabbitMQ 4 ve S3-compatible API üzerinden RustFS içerir. Go module yolu `github.com/Onurlulardan/harpy`, Go sürümü 1.26'dır.

Mevcut repository kapsamı, kaynak kodda implement edilmiş backend yetenekleridir. Forms Engine, Workflow Engine, AI modülü, MCP server ve tarihsel SQL migration dump'ları mevcut kapsamda değildir. Bunları varsayılan ürün yeteneği olarak ekleme.

Setup CLI komutu `go run ./cmd/setup` şeklindedir. Deployment profilini doğrular, yerel configuration ve secret materyalini hazırlar, profil dosyalarını yazar ve sonraki deployment komutlarını gösterir. Docker başlatmaz veya tüm uygulama yaşam döngüsünü tek başına yürütmez.

## Çalışma protokolü

AGENTS, INDEX, SKILL ve ilgili mimari dosyalarını salt okunur keşfetme komutları dışında, kod veya doküman değiştirmeden önce:

1. `skills/INDEX.md` dosyasını oku.
2. İş ile eşleşen `skills/*/SKILL.md` dosyalarını eksiksiz oku.
3. Birden fazla modülü etkileyen işte `impact-analysis` skill'ini de oku.
4. `docs/architecture/` altındaki ilgili mimari metinleri oku.
5. Kapsam veya sahiplik belirsizse değişiklik yapmadan önce etki ve karar sınırını netleştir.

Gerekli skill'ler okunmadan implementasyon değişikliği yapma.

Local API ve worker, `ops/local/compose.runtime.yml` içinde container olarak çalışır. Desteklenen tam local workflow `task dev` ( `local:up` kısa yolu) komutudur. Akış sırasıyla `prepare`, `infra`, `bootstrap`, `migrate`, `installation:artifacts`, `license`, `tenant-executor`, `initialize`, `services` ve `verify` adımlarını yürütür (bkz. `.taskfiles/local.yml`). `license` adımı atlanırsa on-premise API gerekli `billing_license_heads` kaydını bulamaz. Tek başına `go run ./core/cmd/api` çalıştırmak bu akışın yerine geçmez.

## Kurallar

1. Elle yazılan dosyalar `code-organization` skill'ine ve fileguard'a uymalıdır. Satır, CRUD ve dosya adı politikasının kaynağı bu skill'dir.
2. Bir modül başka modülün tablolarına, domain'ine veya adapter'larına erişemez.
3. Handler business logic taşımaz. Repository authorization kararı vermez.
4. Transaction sınırını application command belirler; HTTP middleware belirlemez.
5. Tenant kimliği doğrulanmış actor context'ten gelir; request header kaynak değildir.
6. Aynı process içindeki modüller HTTP veya RabbitMQ ile haberleşmez.
7. RabbitMQ state kaynağı değil, integration transport'tur.
8. Forms Engine, Workflow Engine, AI modülü, MCP server veya tarihsel SQL dump'larını var olan ürün kapsamıymış gibi ekleme.
9. Dynamic query yalnız REST Query `QuerySpec` üzerinden çalışır.
10. Bilinmeyen field, operator veya node sessizce kabul edilmez.
11. Generated dosyalar elle değiştirilmez ve generated provenance header'ını korumalıdır.
12. Kullanıcı açıkça istemedikçe git komutu çalıştırılmaz.
13. Geçici framework, storage, transport veya domain modeli kurulmaz. Migration ayrı bir core sınırıdır; kalıcı state'e, otomatik doğrulanan çıkış koşullarına ve açık sahipliğe sahiptir.

## Modül sınırları

1. Her çalışma kolu tek bir modül veya açıkça tanımlanmış tek bir owner sınırıyla sınırlıdır. Yazılabilir dizinler görevde belirtilir; çakışan yazma alanı verilmez.
2. Modül kapanana kadar yalnızca o modülün source'u, modüle özel HTTP contract'ı, migration'ı, permission implementasyonu ve hedefli testleri değiştirilir. Schema SQL, Harpy için yazılır.
3. Bir modül şu kanıtlar olmadan tamamlanmış sayılmaz:
   - Endpoint, application, adapter ve gerekiyorsa DB routine birbirine bağlıdır.
   - Production constructor, test fixture veya stub yerine gerçek owner API ve concrete adapter kullanır.
   - Aynı HTTP method ve path birden fazla modül tarafından sahiplenilmez.
   - Tenant ve permission kararları hedefli testlerle fail-closed olarak kanıtlanmıştır.
   - Migration sürümleri başka owner ile çakışmaz.
   - Modülün hedefli testleri geçmiştir.
4. Eksik owner API, concrete adapter, permission kararı, DB routine, migration veya hedefli test blocker'dır. Modül bu eksiklerle kapatılmaz.
5. Yeni permission uydurulmaz. Farklı authorization modeli ancak açık kararla ve fail-closed testlerle kabul edilir.
6. Kapanmış modül, hedefli test veya canlı runtime kanıtı zorunlu bir hatayı gösterirse yeniden açılır. Yeni kapsam tek başına yeniden açma gerekçesi değildir.

## Paylaşılan yüzeyler

Modüller kapanana kadar şu yüzeyler kilitlidir: operation manifestleri, authority manifestleri, generated OpenAPI aggregate, global permission kataloğu, migration manifesti, archguard ve fileguard sayaçları.

Bu dosyalar paralel modül çalışma kolları tarafından düzenlenmez. İlgili modüller kapandıktan sonra tek bir owner ortak normalizasyon geçişini yapar. Bir modül yeniden açılırsa ortak aggregate üzerinde art arda düzeltme yapılmaz.

## Kapanış sırası

1. Kapanmış modüllerin hedefli testleri.
2. Operation, authority, permission ve OpenAPI aggregate'inin tek seferde kurulması.
3. Migration manifestinin tek seferde normalizasyonu.
4. `fileguard` ve `archguard`.
5. Veri korumalı migration uygulaması ve local stack doğrulaması.
6. Login, tenant seçimi ve IAM authorization akışı.

Build, unit test veya yeşil statik kontrol tek başına tamamlanma kanıtı değildir. Harpy tamamlandı denmeden önce local stack çalışmalı ve bu akışlar denenmelidir.

Eksik wiring varken sayaç veya snapshot güncelleyerek kalite kapısı kapatılmaz. Login ve localhost kabulü ayrı bir owner geçişidir. Bu geçişin içinde E2E retry döngüsü yoktur. README güncellemesi eksik wiring'i tamamlamaz.

## Sahiplik

- `AGENTS.md`: Zorunlu repository çalışma protokolü.
- `skills/INDEX.md`: İş türü ile skill eşlemesi.
- `skills/*/SKILL.md`: Alana özel prosedürler.
- `docs/architecture/`: Güncel mimari ve modül dokümantasyonu.
- `README.md`: Repository özeti, kurulum yolları, runtime süreçleri ve başlangıç noktaları.

Aynı kuralı birden fazla dosyada tekrarlama. Kuralın sahibi olan dosyaya bağlantı ver.

## Tamamlama

- İlgili skill checklist'leri karşılanmıştır.
- Etkilenen mimaride açık karar kalmamıştır.
- Hedefli testler ve architecture testleri geçmiştir.
- `go run ./tools/fileguard` ve `go run ./tools/archguard` başarılıdır.
- Tenant, permission ve migration etkisi raporlanmıştır.
- Yeni davranışın contract'ı ve observability'si vardır.
