# 11 — Geliştirici başlangıç rehberi

## 1. Ön koşullar

Backend geliştirme için:
- Go 1.26.5 toolchain
- Docker Engine/Desktop ve Compose v2 (local stack için)
- Go Task (task)
- İlk image build'inde internet erişimi

Sister frontend-v3 projesini de çalıştıracaksan:
- Node.js 22+
- Yarn 1
- Sibling frontend repository'sinin local kurulumu

ClamAV signature ilk kurulumunda Docker'a en az 4 GB RAM ayrılması önerilir.

## 2. Repository'yi tanı

Önce şu dosyaları oku:
1. README.md
2. AGENTS.md
3. skills/INDEX.md
4. docs/architecture/README.md
5. Değişiklik yapacağın modülün skill dosyası
6. Modülün mevcut source, test ve contract dosyaları

Skill okumadan dosya değiştirme. Birden fazla modül veya common boundary etkileniyorsa impact-analysis ve modular-architecture yönergelerini uygula.

## 3. Local backend'i başlatma

Repository root'unda:

    task dev

Bu komut task local:up görevine yönlenir. Local Taskfile config, image build, secret projection, dependency startup ve readiness kontrol adımlarını yönetir.

Local stack'i durdurmak:

    task stop

Bu normal durdurma named volume verisini korur. task local:destroy farklıdır; onay sonrası local DB, queue, object ve mail state'ini kalıcı olarak siler.

## 4. Setup CLI

Profil dosyalarını ayrıca oluşturmak için:

    go run ./cmd/setup

Sık kullanılan seçenekler:
- --defaults: varsayılan yanıtlarla çalış.
- --profile single-tenant
- --profile multi-tenant
- --rotate-secrets: yerel secret materyalini yenile.
- --management-listen host:port: management bind address'i açıkça seç.

Setup tamamlandığında oluşturduğu profile ve izleyen process komutlarını terminalde gösterir.

## 5. Modül içinde değişiklik yapma

Her iş için şu soruları cevapla:
1. Business capability owner'ı hangi modül?
2. Public command/query kontratı nedir?
3. Tenant/global transaction scope hangisi?
4. Permission ve actor nereden doğrulanır?
5. DB state, audit, idempotency ve outbox atomicity ihtiyacı var mı?
6. External side effect varsa worker/job sınırı nedir?
7. Contract, test ve architecture guard etkisi nedir?

Mevcut bir modülün katmanını takip et; yalnızca gerçekten kullanılan package'ı ekle. Başka modülün internal package'ına bağlanma.

## 6. Çalışırken kullanışlı komutlar

    # Go testleri
    go test ./...

    # Dosya organizasyonu
    go run ./tools/fileguard

    # Import/architecture kuralları
    go run ./tools/archguard

    # Local stack durum/operasyon
    task local:logs
    task local:logs:api
    task local:logs:worker
    task local:check

## 7. Sorun giderme

**API başlamıyor:** API management readiness ve loglardan dependency preflight, DB role/schema contract, IAM runtime ve object/KMS readiness'i kontrol et.

**Worker başlamıyor:** PostgreSQL ve RabbitMQ health'i, producer/consumer catalog compatibility, processor registry ve worker admission durumunu incele.

**Bir endpoint 403 dönüyor:** Actor context, tenant scope, IAM action, resource-level policy ve field permission'ı birbirinden ayır.

**Bir job ilerlemiyor:** DB job state, attempt, lease owner/expiry, last error ve worker processor registration'ı incele. RabbitMQ'yu job state kaynağı gibi ele alma.

**Belge ready olmuyor:** Upload state, digest/byte count, MIME/signature, scanner freshness, S3 object version/receipt ve cleanup requirement alanlarını kontrol et.

## 8. Kaynak kodda gezinme

| Konu | Dizin |
| --- | --- |
| API entry + route wiring | core/cmd/api, core/internal/composition/api |
| Worker/job wiring | core/cmd/worker, core/internal/composition/worker |
| Operator | core/cmd/operator, core/internal/composition/operator |
| İş modülleri | core/internal/modules/ |
| Platform | core/internal/platform/ |
| Capabilities | core/internal/capabilities/ |
| HTTP/event/runtime contracts | contracts/ |
| Test ve guard araçları | tools/ |
| Local stack ve Taskfile | ops/local/, .taskfiles/ |

## 9. Değişiklik teslim checklist'i

- İlgili skill checklist'i tamamlandı.
- Source değişikliği doğru owner sınırında.
- HTTP/event contract'ları tutarlı.
- Unit ve targeted integration/security testleri eklendi.
- go test ./..., fileguard ve archguard sonuçları kontrol edildi.
- Local runtime gerekiyorsa API ve worker readiness doğrulandı.
- PR açıklaması yapılan değişikliği, testleri ve bilinen kısıtları belirtir.
