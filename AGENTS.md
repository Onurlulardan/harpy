# Harpy

Harpy, adını harpi kartalından (harpy eagle) alır. Go ile yazılmış bir backend starter'dır: IAM, iş kuyruğu, worker, single-tenant ve multi-tenant, platform, composition ve `core/cmd/api`, `core/cmd/worker`, `core/cmd/migrator`, `core/cmd/operator`, `cmd/setup`. Yerel üçüncü parti Postgres 18, RabbitMQ ve RustFS'tir. Blob erişimi tek S3 API'dir (`aws-sdk-go-v2`).

`forms`, `workflow`, AI modülü, MCP sunucusu ve tarihsel SQL migration dökümleri starter'da yoktur. Migrator kodu starter'dadır. Şema SQL'i Harpy için yazılır. Bu parçalar eklenmez.

İlk komut `go run ./cmd/setup`. Setup profil sorar, Docker başlatmaz, compose / migrate / api / worker komutlarını basar.

Go module yolu `github.com/Onurlulardan/harpy`. Go sürümü `1.26`.

## Çalışma protokolü

AGENTS, INDEX ve SKILL okumaya yarayan salt okunur keşif komutları dışında, kod veya doküman değiştirmeden önce:

1. `skills/INDEX.md` dosyasını oku.
2. İş ile eşleşen `skills/*/SKILL.md` dosyalarını eksiksiz oku.
3. Birden fazla modülü etkileyen işte `impact-analysis` skill'ini de oku.
4. İlgili mimari metni `docs/` altından oku.
5. Belirsizlik varsa kod yazma; önce etki ve karar sınırını netleştir.

Skill okumadan yapılan değişiklik hatalıdır.

Yerel API ve worker `ops/local/compose.runtime.yml` içinde container'dır. Komut `task dev` (`local:up`). `task dev` kullanılmazsa aynı sıra eksiksiz tamamlanır: `prepare`, `infra`, `bootstrap`, `migrate`, `installation:artifacts`, `license`, `tenant-executor`, `initialize`, `services`, `verify` (`.taskfiles/local.yml` `up`). `license` atlanırsa on-premise açılış `billing_license_heads` satırını bulamaz ve API kalkmaz. `go run ./core/cmd/api` bu sıranın yerine geçmez.

## Kurallar

1. Elle yazılan dosyalar `code-organization` skill'ini ve fileguard'ı geçer. Satır, CRUD ve dosya adı politikasının tek kaynağı o skill'dir.
2. Bir modül başka modülün tablosuna, domain'ine veya adapter'ına erişemez.
3. Handler business logic taşımaz. Repository authorization kararı vermez.
4. Transaction sınırını application command belirler. HTTP middleware belirlemez.
5. Tenant kimliği doğrulanmış actor context'ten gelir. Header kaynak değildir.
6. Aynı process içindeki modüller HTTP veya RabbitMQ ile konuşmaz.
7. RabbitMQ state kaynağı değil, integration transport'udur.
8. `forms`, `workflow`, AI modülü, MCP sunucusu ve tarihsel SQL dökümü eklenmez.
9. Dynamic query yalnız REST Query `QuerySpec` üzerinden çalışır.
10. Bilinmeyen alan, operator veya node sessizce kabul edilmez.
11. Generated dosya elle değiştirilmez. Generated header taşımak zorundadır.
12. Kullanıcı istemedikçe git komutu çalıştırılmaz.
13. Geçici framework, storage, transport veya domain modeli kurulmaz. Migration mekanizması çekirdekten ayrı bir sınırdır; state'i kalıcıdır, çıkış koşulu otomatik doğrulanır ve sahibi bellidir.

## Modül sınırı

1. Her çalışma kolu tek bir modül ya da açıkça tanımlanmış tek bir owner sınırı ile sınırlanır. Yazabileceği dizinler görevde yazılır. Çakışan yazma alanı verilmez.
2. Modül kapanana kadar yalnız o modülün source'u, modüle özel HTTP kontratı, migration'ı, permission uygulaması ve hedefli testleri değişir. Şema SQL'i Harpy için yazılır.
3. Bir modül şu kanıtlar olmadan kapanmaz:
   - Endpoint, application, adapter ve gerekiyorsa DB rutini bağlıdır.
   - Production constructor, test fixture ve stub olmadan gerçek owner API ve concrete adapter ile kurulabilir.
   - Aynı HTTP method ve path iki modül tarafından sahiplenilmez.
   - Tenant ve permission kararı hedefli testlerle fail-closed kanıtlanmıştır.
   - Migration sürümleri başka owner ile çakışmaz.
   - Modülün hedefli testleri geçmiştir.
4. Eksik owner API, concrete adapter, permission kararı, DB rutini, migration veya hedefli test dış blocker sayılmaz. Modül bu eksikle kapanmaz.
5. Yeni bir permission uydurulmaz. Farklı yetki modeli ancak fail-closed test ve açık kararla kabul edilir.
6. Kapanmış modül, hedefli test veya canlı runtime kanıtı zorunlu bir hata gösterirse yeniden açılır. Yeni kapsam, kapanmış modülü açma gerekçesi değildir.

## Ortak yüzey

Modül kapanmadan önce şu yüzeyler kilitlidir: operation manifestleri, authority manifestleri, generated OpenAPI aggregate, global permission kataloğu, migration manifesti, archguard ve fileguard sayaçları.

Bu dosyaları paralel kollar düzenlemez. İlgili modüller kapandıktan sonra tek sahip bir geçişte değiştirir. Bir modül yeniden açılırsa ortak yüzeyde art arda düzeltme yapılmaz.

## Kapanış sırası

1. Kapanmış modüllerin hedefli testleri
2. Operation, authority, permission ve OpenAPI aggregate'inin tek kurulumu
3. Migration manifestinin tek normalizasyonu
4. `fileguard` ve `archguard`
5. Veri korumalı migration uygulaması ve local stack
6. Login, tenant seçimi ve IAM authorization akışı

Build, unit test veya yeşil statik kapı tek başına tamamlanma kanıtı değildir. Local stack kalkmadan ve bu akışlar denenmeden Harpy tamamlandı denemez.

Eksik wiring varken sayaç veya snapshot güncelleyerek kapı kapatılmaz. Login ve localhost kabulü ayrı owner geçişidir. Bu geçişin içinde E2E tekrar döngüsü yoktur. README güncellemesi eksik wiring'i kapatmaz.

## Sahiplik

- `AGENTS.md`: zorunlu protokol.
- `skills/INDEX.md`: iş türü ile skill eşlemesi.
- `skills/*/SKILL.md`: alana özel prosedür.
- `docs/harpy-starter/`: starter şekli, ilk çalıştırma ve guard sözleşmesi.
- `README.md`: klon, setup, compose, migrate, api ve worker sırası.

Aynı kuralı birden fazla dosyada yeniden tanımlama. Sahip dosyaya link ver.

## Tamamlama

- İlgili skill checklist'leri karşılanmıştır.
- Etkilenen mimaride açık karar yoktur.
- Targeted testler ve architecture testleri geçmiştir.
- `go run ./tools/fileguard` ve `go run ./tools/archguard` başarılıdır.
- Tenant, permission ve migration etkisi raporlanmıştır.
- Yeni davranışın kontratı ve gözlemlenebilirliği vardır.
