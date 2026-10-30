# 10 — Test ve mühendislik standartları

## 1. CI görünümü

CI workflow .github/workflows/ci.yml içindedir. push ve pull_request olaylarında çalışır. Workflow'da üç ana job vardır:

| Job | Kontrol |
| --- | --- |
| unit | Go test suite |
| fileguard | Dosya organizasyonu ve generated provenance kuralları |
| archguard | Paket dependency ve architecture kuralları |

Unit job Go 1.26.5 toolchain kullanır ve go test ./... çalıştırır. Fileguard ve archguard ayrı CI job'larıdır; böylece kod testlerinden bağımsız mimari ihlaller de PR kapısında yakalanır.

## 2. Test türleri

| Tür | Ne zaman kullanılır | Örnek alan |
| --- | --- | --- |
| Unit | Saf domain behavior ve küçük use-case | State transition, parser, money math |
| Application | Port fixture ile use-case davranışı | Authorization, idempotency, transaction orchestration |
| Adapter contract | DB/HTTP/provider contract | Error mapping, schema, wire payload |
| PostgreSQL integration | Gerçek constraint/transaction/session davranışı | Tenant scope, role preflight, row locking |
| Broker integration | Broker protocol ve confirmation | Rabbit publish/consume, reconnect |
| Security negative | Yetkisiz/bozuk/başka tenant girdisi | IDOR, token tampering, role escalation |
| Property/fuzz | Geniş input space ve invariant | Query parser, canonical JSON, crypto framing |
| Architecture | Import yönü ve owner boundary | Cross-module import/SQL prohibition |

Harpy'de test yoğunluğu özellikle IAM, Billing, Documents, Analytics ve async runtime tarafında yüksektir. Test isimlerinin yanında ilgili source dosyalarını ve fixture contract'larını da incele.

## 3. Local kalite komutları

Repository root'unda:

    go test ./...
    go run ./tools/fileguard
    go run ./tools/archguard
    task check

Local Taskfile ek olarak local config, tools ve operational utility testlerini çalıştırır. Local stack doğrulaması için task local:check tanımlıdır.

## 4. Fileguard standardı

tools/fileguard elle yazılan dosyalarda:
- En fazla 200 fiziksel satır,
- UTF-8 text,
- Symlink olmaması,
- Genel amaçlı dumping-ground dosya adlarının yasak olması,
- Generated dosyalarda provenance/header doğrulaması
kurallarını uygular.

Lockfile ve doğrulanmış generated output için tanımlı istisnalar vardır. Yeni istisna eklemek dosya sınırını aşmanın yolu değildir.

İzin verilmeyen genel dosya adları:
crud.go, handler.go, service.go, repository.go, utils.go, helpers.go, common.go, misc.go.

## 5. Archguard standardı

tools/archguard import graph üzerinde mimari kısıtları denetler. Örnek kontroller:
- Platform business modules import edemez.
- Process entrypoint'leri yalnız kendi composition girişini import edebilir.
- Cross-module import yalnız hedef modülün api paketine yapılabilir.
- Domain HTTP, database driver veya cloud SDK'ya bağlanamaz.
- Application adapter/framework detayına bağlanamaz.
- AWS S3 SDK yalnız blob/object adapter alanında kullanılabilir.
- MinIO/RustFS SDK bağımlılıkları business code içine sızamaz.

## 6. Dosya organizasyonu

- Her dosyanın bir primary responsibility'si olmalıdır.
- CRUD ve use-case operasyonları ayrı dosyalara bölünür.
- get ve list ayrıdır.
- Routes dosyası yalnız route registration yapar.
- DTO, domain entity ve DB row modeli tek struct altında birleştirilmez.
- Consumer'a yakın interface tercih edilir.
- Generic helper yerine domain semantiği taşıyan adlandırma kullanılır.

## 7. Test yazma ilkeleri

- Zaman, ID ve random source testte kontrol edilebilir dependency olmalıdır.
- Hata ve retry testleri transient/permanent/outcome-unknown farkını kapsamalıdır.
- Negative authorization testleri en az positive test kadar önemlidir.
- Tenant testleri yalnız tenant ID eşitliğine değil DB schema/session sınırına bakmalıdır.
- Event/job testleri duplicate, reorder, stale lease, crash ve retry durumlarını kapsamalıdır.
- Crypto testleri tamper, invalid framing, wrong key/context ve limit sınırını doğrulamalıdır.
- Integration test gerçek PostgreSQL/Broker davranışına ihtiyaç duyuyorsa mock ile geçiştirilmemelidir.

## 8. PR review checklist

- Değişen business state'in owner modülü doğru mu?
- HTTP contract ve event contract ayrı ayrı güncellendi mi?
- Tenant/permission/step-up kararı fail-closed mu?
- Cross-module direct DB access var mı?
- Idempotency, version conflict ve concurrent request test edildi mi?
- Secret, token, file content ve PII log/event payload'ından çıkarıldı mı?
- Local runtime composition gerçekten gerekli dependency'lerle kurulabiliyor mu?
- Targeted test, full unit suite ve iki architecture guard sonucu raporlandı mı?

## 9. Normatif mühendislik kaynakları

- AGENTS.md: değişiklik öncesi repository protokolü.
- skills/INDEX.md: gerekli skill eşlemesi.
- skills/code-organization/SKILL.md: file/line/name standardı.
- skills/modular-architecture/SKILL.md: modül ve dependency sınırı.
- skills/testing-quality/SKILL.md: test stratejisi.
- tools/fileguard/, tools/archguard/: otomatik kontrol implementasyonları.
