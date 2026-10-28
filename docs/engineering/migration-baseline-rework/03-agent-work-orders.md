# 03 — Agent iş emirleri, dosya sınırları ve bağımlılıklar

## Genel talimatlar (tüm agent'lar için zorunlu)

1. Önce bu klasördeki dört planı ve belirtilen kaynak kodları oku. README'ye bakıp tahminle refactor yapma.
2. Bu çalışma **backwards-compatible migration upgrade çalışması değildir**. Amaç yeni starter repo'sunun başlangıç şemasını son geçerli hâle getirmek ve legacy migration zincirini kaldırmaktır.
3. Eski SQL'leri basitçe silme; önce temiz PostgreSQL 18 üzerinde final schema baseline'ını üret, kaynak/target katalog karşılaştırmasını ve runtime smoke testlerini geçir.
4. Kullanıcının local named volume'larını/gerçek DB'sini silme. Her destructive test disposable Compose project, temporary Postgres container veya disposable volume kullansın. `task local:destroy`, volume prune veya DB reset komutları otomatik çalıştırılmasın.
5. Hash/digest/UUID/license/seed değerleri uydurma. Migration hash manifesti, release archive veya custom store'u yeni adlarla yeniden yaratma.
6. Eski sistemdeki her `core/internal/platform/migrator` ve `tools/localrelease` dosyasını topluca silme. Call graph ve tests ile hangi davranışın yalnız legacy migration framework'e ait olduğunu ispatla. License, role, extension, secret, installation, backup/restore, tenant identity ve RLS davranışları ayrı domain sorumluluğudur.
7. Bir agent aynı anda başka agent'ın sahip olduğu dosya grubunu değiştirmesin. Fazlar sequential merge/commit mantığıyla ilerlesin; merge sonrası testler yeniden çalıştırılsın.
8. Her fazın sonunda `git diff --check`, ilgili tests ve değişen dosyaların listesi raporlansın. Test gevşetme, ignored test veya `go test` atlama kabul edilmez.

## Faz 0 — Audit ve source-of-truth doğrulama (read-only agent)

**Dokunulacak dosya yok; yalnız envanter/rapor.**

- `internal/migrationbundle/**`, `core/internal/platform/migration/**`, `core/internal/platform/migrator/**`, `core/internal/composition/{migrator,operator}/**`, `cmd/setup/**`, `tools/localrelease/**`, `tools/internal/migrationmanifest/**`, Dockerfile/Compose/Taskfile/CI çağrı noktalarını tara.
- Ekli local logdaki `history version 20260722050009 differs from verified metadata` hatasının temiz DB'de de tekrarlanıp tekrarlanmadığını kaydet.
- Boş PostgreSQL 18'e current main'i deploy edip platform + örnek tenant final şemasını kur. Temiz kurulum başarısızsa hangi migration, SQLSTATE, DB object veya metadata kontrolünde durduğunu tam olarak raporla. Başarısız zincirden yarım dump üretme.
- Tüm gerekli database object/seed/extension/role/privilege'leri envanterle. Runtime-generated records ile statik başlangıç verilerini ayır.
- Çıkış: baseline source DB başarılı; kaynak final schema envanteri, gerekli seed listesi ve tüm çağrı noktaları raporlanmış.

## Faz 1 — Baseline üretimi (database/schema agent)

**Sahiplik:** yalnız yeni baseline SQL dosyaları ve bunları tekrarlanabilir şekilde üretmeye/test etmeye yarayan küçük script/test. Henüz eski dosyaları silme.

- Platform ve tenant için gerekli kapsamları aynı baseline release'i içinde üret.
- Başlangıç nesnelerini pg_dump üzerinden al; dump'ı elle yazmaya başlama. Role/extension/grant ve gerekli data farklarını açıkça ele al.
- Baseline'dan boş DB ve yeni tenant schema oluşturup source final schema ile katalog karşılaştırmasını çalıştır.
- On-premise kurulumun gerekli installation manifest ve billing license/offering kayıtları setup/operator authority akışıyla oluşabildiğini doğrula; signing key/license gibi runtime verileri baseline'a gömme.
- Çıkış: fresh platform DB + fresh tenant schema testleri geçiyor, object/seed farkları açıklanmış ve baseline dosyaları sabitlenmiş.

## Faz 2 — Goose runtime sadeleştirmesi (Go migration agent)

**Sahiplik:** migration runner/provider/history ve migration-specific contracts; build wiring/old SQL deletion yok.

- `build_goose_provider.go` ve `build_tenant_goose_provider.go` içindeki `DialectCustom` + custom `WithStore` yolunu Goose native PostgreSQL store'a geçir.
- Baseline'ı bir kere uygulat; yalnız sonraki timestamp'li SQL'leri Goose ile sırayla uygulat. Platform/tenant scope ve paralel tenant initialization lock gereksinimini kanıtla.
- `classify_history.go`, history supersession, custom `Store`/`TenantStore` ve selected migration metadata bağımlılıklarını native Goose history ile değiştir.
- Goose native store dinamik tenant schema history'sini desteklemezse minimal bir store/adapter yap ve gerekçesini tek dokümanda yaz; eski metadata modelini koruma.
- `core/internal/composition/operator/startup.go` gibi yalnız schema contract isteyen call site'larda `FreshMigrationDocuments()` ile corpus toplamayı kaldır. Migrator ancak migration uygulayacağı için kendi kapsamındaki dosyaları yükler.
- Çıkış: tests prove baseline up, second up is no-op, new migration applies once, invalid order fails and per-tenant history is isolated.

## Faz 3 — Bootstrap/build/local dev sadeleştirmesi (platform tooling agent)

**Sahiplik:** `Taskfile.yml`, `.taskfiles/**`, `ops/local/Dockerfile`, `ops/production/Dockerfile`, `ops/local/compose*.yml`, `tools/localrelease/**`, `cmd/setup/**`, `.github/workflows/ci.yml`.

- `foundation-contract`/`foundation-applied`/hard-coded 138-entry manifest ve separate full release aşamalarını tek initial baseline/release akışına indir.
- Docker local ve production build'i eski manifest generation ve migration-release ZIP üretiminden çıkar.
- `task dev` clean volume üzerinde tek anlaşılır bootstrap → platform baseline → installation/license/setup → tenant baseline → API/worker akışını başlatmalı. Başlangıç sırasını dependency'lere göre belirle; runtime secret veya signed license'ı SQL baseline'a taşıma.
- `tools/localrelease` içindeki migration artifact kodlarını ayır; lisans, key, grant, extension, setup ve recovery kodları gerekiyorsa kalmalı.
- Çıkış: build/setup tek authoritative baseline kullanıyor; hiç kimse eski custom manifest dosyası/hash'i elle kopyalamıyor.

## Faz 4 — Legacy migration framework removal agent

**Başlangıç koşulu:** Faz 1–3 tamamlandı, baseline ve Goose testleri yeşil.

- Eski 2.000 SQL dosyası ve supersession JSON'larını kaldır; yeni baseline ve bundan sonraki migrations hariç legacy `.fresh.sql` zinciri kalmasın.
- `internal/migrationbundle/**`, `tools/internal/migrationmanifest/**`, 109 historical release manifest'i, `release-index.json`, migration release archive/packer, yalnız legacy chain'e ait registry/select/manifest attestation code ve contract alanlarını kaldır.
- Tüm Go embed pattern'lerini, CI/Taskfile/Dockerfile references, tests, FileGuard/ArchGuard ve docs'u güncelle.
- `tools/localrelease` veya `core/internal/platform/migrator` klasörünü körlemesine kaldırma; yalnız call graph'ında artık kullanılmayan migration-specific kodu sil.
- Çıkış: `rg` taramasıyla stale reference yok; starter repository yalnız baseline + future migration'larla çalışıyor.

## Faz 5 — Bağımsız doğrulama agent'ı

Bu agent diğer fazların yazdığı kodu yeniden tasarlamaz; test ve diff inceler.

- Clean PostgreSQL 18 platform install ve fresh tenant creation.
- Re-run bootstrap idempotency ve API/worker readiness.
- Gelecek migration senaryosu: yeni migration ekle → Goose uygulasın → ikinci `up` değişiklik yapmasın → version table doğru sürümü göstersin.
- SQL geçmişi koruma kuralı: önceki commit'teki migration dosyasına edit/delete gelirse CI fail; new `.sql` eklenmesi geçsin.
- Roles, ownership, grants, extensions, constraints, functions, triggers, RLS/policies, tenant isolation, license offering, backup/restore ve setup idempotency regression testleri.
- Çıkış: test command, exit code ve önemli schema comparison sonuçları raporlanmış.

## Agent'lar arası sıralama

Faz 0 → Faz 1 → Faz 2 → Faz 3 → Faz 4 → Faz 5. Bunları migration kaynakları ve build wiring üzerinde paralel çalıştırma. Tek bir entegrasyon branch'i üzerinden ilerle; her aşama kabul edilmeden sonraki agent'a eski katmanları sildirme.
