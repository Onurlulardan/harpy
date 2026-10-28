# Migration sistemi: cerrahi sadeleştirme ve tek başlangıç baseline'ı

## Karar

Harpy yeni bir starter olarak temiz başlayacak. Eski migration zincirini devam ettirme veya eski veritabanlarına uyumluluk sağlama hedefi yok. Git çalışma ağacında yalnızca güncel başlangıç şeması ve bundan sonra yazılacak migration'lar kalacak.

**Hedef akış:** boş PostgreSQL → güncel şemayı kuran tek baseline release'i → Goose history → bundan sonraki her değişiklik için yeni SQL migration. Eski 1.982 seçilmiş migration yeni kurulumda çalıştırılmayacak; eski migration dosyaları baseline doğrulanıp geçiş tamamlandıktan sonra mevcut çalışma ağacından kaldırılacak. Git'in geçmiş commit'leri ayrıca yeniden yazılmayacak.

### “Tek baseline” ne demek?

Tek bir başlangıç sürümü, tek bir üretim/kurulum akışı ve tek bir schema snapshot kaynağı. Kodda platform veritabanı ile her yeni tenant için ayrı schema/migration yaşam döngüsü bulunduğundan, doğrulama bunu gerektiriyorsa aynı baseline release'i **platform** ve **tenant** için iki SQL artifact'ı üretebilir. Bu, iki ayrı foundation/full release sistemi kurmak için gerekçe değildir. Tenant ve platform gerçekten aynı lifecycle'da değilse her şeyi tek SQL dosyasına sıkıştırıp tenant provisioning'i bozma.

## Yeni mimaride ne kalacak?

- PostgreSQL 18 ve `github.com/pressly/goose/v3` mevcut bağımlılığı.
- Goose'un native PostgreSQL version store/history'si; sıra dışı migration uygulaması kapalı.
- Bir adet ilk baseline migration seti; sonraki değişiklikler timestamp'li yeni `.sql` dosyaları.
- Platform ve tenant migration'larını doğru veritabanı/schema kapsamına uygulayan ince entegrasyon kodu.
- Gerekli roller, grants, extensions, RLS, tenant izolasyonu, installation/license bootstrap ve recovery işlevleri. Bunlar migration registry'ye bağlı olmak zorunda değil.
- CI'da geçmiş migration dosyalarının sonradan değiştirilmesini engelleyen basit kontrol ve boş PostgreSQL'de gerçek kurulum testi.

## Neler hedef dışı?

- Eski migration geçmişini yeni starter'da sürdürmek.
- Her SQL dosyasını içeren custom manifest, digest, JCS serialization, release archive/index ve supersession katmanlarını başka isimlerle yeniden kurmak.
- Her migration için owner/scope/generation/run ID kaydeden özel Goose Store'u, yalnızca migration çalıştırmak için yaşatmak.
- Hash'i elle düzeltmek, yalnız testi yeşile çevirmek veya 1/0 assertion'larını tüm repo genelinde toplu değiştirmek.
- Goose'u değiştirmek ya da ikinci bir migration motoru eklemek.

## Belgeler

- [01 — Mevcut kod envanteri ve kaldırma haritası](01-current-state-and-cut-list.md)
- [02 — Baseline üretimi ve hedef Goose mimarisi](02-baseline-and-goose.md)
- [03 — Agent iş emirleri ve paralelleştirme sınırları](03-agent-work-orders.md)
- [04 — Kabul kriterleri, testler ve local recovery](04-acceptance-and-recovery.md)

## Gözden geçirilen mevcut durum

İnceleme GitHub `main` başlığı `0a085c1f08d83e80ae8b1c13f63daacd8293de87` üzerinden yapıldı. Bu commit, önceki roadmap Markdown dosyalarının silindiği `delete docs` commit'idir.

- Repo ağacında migration dizinleri altında 2.000 SQL dosyası var; mevcut fresh manifest 1.982 seçilmiş entry içeriyor: 969 platform, 1.013 tenant.
- `core/internal/platform/migration/release/` altında 109 historical migration manifest'i var.
- Goose bağımlılığı zaten `go.mod` içinde (`pressly/goose/v3 v3.27.2`), ancak custom `Store`, custom history tablosu, release registry ve manifest sistemi Goose'un etrafına ikinci bir migration framework'ü koyuyor.
- Local logunda `core-migrator deploy up` şu hatayla duruyor: `history version 20260722050009 differs from verified metadata`. Kod bu mesajı `core/internal/platform/migrator/classify_history.go` içinde üretiyor.

Önemli: baseline, eski SQL dosyalarından rastgele seçilerek oluşturulmayacak. Önce temiz, geçici bir PostgreSQL 18 ortamında mevcut son geçerli şema oluşturulacak; sonra bu şema doğrulanmış başlangıç baseline'ına dönüştürülecek.
