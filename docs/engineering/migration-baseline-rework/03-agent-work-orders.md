# 03 — Agent iş emirleri, dosya sınırları ve bağımlılıklar

## Genel talimatlar

1. Önce bu klasördeki planların tamamını oku. README'ye bakıp tahminle refactor yapma. Doğrulanmış boşluklar [05](05-verified-gaps.md) içindedir.
2. Bu çalışma backwards-compatible migration upgrade çalışması değildir. Amaç yeni starter'ın başlangıç şemasını son geçerli hâle getirmek ve legacy migration zincirini kaldırmaktır.
3. Eski SQL'leri silerek başlama. Temiz PostgreSQL 18 üzerinde final schema baseline'ını üret, katalog karşılaştırmasını ve runtime smoke testlerini geçir. `pg_dump` çıktısını Goose dosyasına sarmak [02](02-baseline-and-goose.md) bölümündedir.
4. Kullanıcının local named volume'larını ve gerçek DB'sini silme. Destructive test disposable Compose project, temporary Postgres container veya disposable volume kullanır. `task local:destroy`, volume prune ve DB reset komutları otomatik çalıştırılmaz.
5. Hash, digest, UUID, license ve seed uydurma. Migration hash manifesti, release archive veya custom store yeni adlarla kurulmaz.
6. `core/internal/platform/migrator` ve `tools/localrelease` topluca silinmez. Call graph ve testler, legacy migration framework ile license, role, extension, secret, installation, backup/restore, tenant identity ve RLS davranışını ayırır.
7. Bir agent başka agent'ın sahip olduğu dosya grubunu değiştirmez. Fazlar tek entegrasyon dalında sırayla ilerler. Her faz kabul edilmeden sonraki agent eski katmanları silmez.
8. Her fazın sonunda `git diff --check`, ilgili testler ve değişen dosyaların listesi raporlanır. Test gevşetme, ignored test ve `go test` atlama kabul edilmez.
9. Catalog business module removal bu branch içinde yürütülür. Generic migration catalog, PostgreSQL catalog, REST Query field catalog, IAM permission/resource catalog ve async/job catalog korunur.

## Önkoşul

Faz 0, Catalog verisi için retain/export/drop kararını zorunlu tutar. Mevcut baseline snapshot'ı tarihsel Catalog DDL içerir; post-baseline cleanup migration yalnız boş/default seed nesnelerini kaldırır ve beklenmeyen dolu Catalog tablolarında hata vererek durur. Baseline dump'ın kendisi yeniden üretilmez. Cleanup'ın bu dump üzerinde çalıştığı `TestBaselineDumpCleanupPostgres18` ile doğrulanır.

Fresh install'da Catalog tabloları ve IAM/Analytics Catalog owner tabloları post-baseline cleanup ile kaldırılır. Cleanup, bilinen boş/default menu seed satırları dışındaki veriyi silmez; beklenmeyen veri görürse transaction hata ile durur. IAM Catalog resource/permission satırları audit ve FK bütünlüğü için retired olarak tutulur; aktif binding'ler revoked yapılır.

## Faz 0 — Audit

**Sahiplik:** yok. Yalnız rapor.

- `internal/migrationbundle/**`, `core/internal/platform/migration/**`, `core/internal/platform/migrator/**`, `core/internal/composition/{migrator,operator}/**`, `cmd/setup/**`, `tools/localrelease/**`, `tools/internal/migrationmanifest/**`, Dockerfile, Compose, Taskfile ve CI çağrı noktalarını tara.
- `history version 20260722050009 differs from verified metadata` hatasının temiz disposable PostgreSQL 18'de de çıkıp çıkmadığını kaydet. Temiz kurulum durursa migration, SQLSTATE, DB object ve metadata kontrolünü yaz. Yarım dump üretme.
- Sembolleri iki listeye ayır. Legacy metadata: `classify_history.go`, supersession, custom `Store` / `TenantStore` içindeki path, SHA, owner, scope, origin, generation. Operasyonel durum: advisory lock, `recoverFinalizedRelease`, `bootstrapRunLifecycle`, `ensureRunRunning`, `finalizeSuccess`, `finalizeFailure`, backup ve release finalization.
- `rg` ile `FreshMigrationDocuments`, `platform_migration_history` ve 1.982 entry assert eden dosyaların tam listesini çıkar. Migrator ağacının içinde ve dışında diye böl.
- Statik seed ile runtime identity, credential, license ve signing key kayıtlarını ayır.
- Çıkış: kaynak şema envanteri, seed listesi, retain listesi, sembol ayrımı ve çağrı listesi. Bu rapor baseline'ı dondurmaz.

## Faz 1 — Baseline

**Başlangıç:** Catalog cleanup migration'ları var; non-empty data için fail-closed preflight mevcut. Baseline dump yeniden üretilmez; cleanup'ın mevcut dump üzerinde çalıştığı PostgreSQL 18 testi vardır.

**Sahiplik:** yeni baseline SQL dosyaları ve onları üreten küçük script ile test. Eski `.fresh.sql` durur.

- Platform ve tenant kapsamlarını aynı baseline release'inde üret. Lifecycle ayrıysa iki SQL dosyası olur.
- Nesneler `pg_dump` ile alınır. Ham dump kaydedilmez. `\restrict` ve `\unrestrict` alınmaz. Her dosya `-- +goose Up` altında `StatementBegin` / `StatementEnd` ile tek statement olur.
- Boş DB ve yeni tenant schema, kaynak final şema ile katalog karşılaştırmasından geçer.
- Installation manifest ve billing license/offering, setup/operator ile oluşur. Signing key ve license baseline'a gömülmez.
- Çıkış: fresh platform DB ve fresh tenant schema testleri geçer. Object ve seed farkları yazılır. Baseline dosyaları sabitlenir.

## Faz 2a — Goose runtime

**Sahiplik:** `core/internal/platform/migrator/**`, `core/internal/composition/migrator/**`, `core/internal/composition/operator/startup.go`. Eski SQL silinmez. Dockerfile, Taskfile ve `tools/localrelease` bu faza ait değildir.

- `DialectCustom` ve custom `WithStore` native PostgreSQL store'a döner.
- Baseline bir kez uygulanır. Sonraki timestamp'li SQL sırayla uygulanır. Out-of-order kapalıdır.
- Native store tenant schema history'sini karşılayamazsa yalnız bu eksik için küçük adapter yazılır. Owner, scope, generation, SHA ve supersession geri gelmez.
- Bir kilit mekanizması kalır. Advisory lock gerekiyorsa Goose locker ile yan yana ikinci kilit kurulmaz.
- Faz 0 ayrımı uygulanır. Legacy release generation'a ait run lifecycle kalkar. Backup veya installation hâlâ kullanıyorsa durum `platform.platform_migration_history` dışına taşınır.
- `startup.go` schema contract için `FreshMigrationDocuments()` ile corpus toplamaz.
- Çıkış: bu ağaçtaki testler baseline up, ikinci up no-op, yeni migration bir kez, geçersiz sıra reddi ve tenant history izolasyonunu gösterir.

## Faz 2b — Migrator dışı call site ve testler

**Başlangıç:** Faz 2a kabul edilmiş. Dosya listesi Faz 0 `rg` raporudur.

**Sahiplik:** migration runner ve baseline wiring. Catalog module kaynakları ile Catalog'a ait IAM/Analytics owner SQL'i bu branch'te kaldırılmıştır; generic migration catalog altyapısına dokunulmaz.

- `platform_migration_history` insert/assert ve 1.982 entry sayımı, native Goose history ile uyumlu hale gelir.
- Çıkış: hedefli testler geçer. Kalan her eski referans gerekçesiyle yazılır.

## Faz 3 — Build ve local

**Sahiplik:** `Taskfile.yml`, `.taskfiles/**`, `ops/local/Dockerfile`, `ops/production/Dockerfile`, `ops/local/compose*.yml`, `cmd/setup/**` içindeki migration manifest yolu, `.github/workflows/ci.yml`.

- `foundation-contract`, `foundation-applied`, sabit 138-entry manifest ve ayrı full release aşaması tek baseline akışına iner.
- Docker build manifest generation ve migration-release ZIP üretiminden çıkar.
- `task dev` temiz volume'da bootstrap, platform baseline, installation/license/setup, tenant baseline, API ve worker sırasını başlatır. Secret ve signed license SQL baseline'a taşınmaz.
- `tools/localrelease` içindeki generic release ve migration artifact üretimi korunur. Catalog package-set, keyring ve deploy-evidence doğrulaması kaldırılır; license, key, grant, extension, setup ve recovery akışları korunur.
- CI, önceki commit'teki `.sql` migration dosyasının değişmesini ve silinmesini reddeder. Yeni dosya eklenmesi geçer.
- Çıkış: build ve setup tek baseline kullanır. Elle kopyalanan manifest hash'i kalmaz.

## Faz 4 — Legacy silme

**Başlangıç:** Faz 1, 2a, 2b ve 3 yeşil.

- Eski 2.000 SQL ve yedi `migration-supersessions.json` kalkar. Baseline ve ondan sonraki migration kalır.
- `internal/migrationbundle/**`, `tools/internal/migrationmanifest/**`, 109 historical release manifest, `release-index.json`, release archive/packer ve yalnız legacy zincire ait registry/select/manifest attestation kalkar.
- Embed, CI, Taskfile, Dockerfile, test, FileGuard, ArchGuard ve bu klasördeki artık referanslar güncellenir.
- `skills/data-tenancy/SKILL.md` migration kuralı bu sete bağlanır. Checksum, owner-module ve eski snapshot upgrade cümleleri starter baseline'ından sonra kalkar. `docs/adr/0006-tenancy-migrations.md` uydurulmaz.
- `tools/localrelease` ve `core/internal/platform/migrator` klasörleri bütün olarak silinmez. Call graph'ında kullanılmayan migration-specific kod silinir.
- Çıkış: `rg` taramasında gerekçesiz eski referans yoktur. Starter baseline ve sonraki migration'larla çalışır.

## Faz 5 — Doğrulama

Bu agent kodun tasarımını değiştirmez. [04](04-acceptance-and-recovery.md) kabul listesini çalıştırır.

- Temiz PostgreSQL 18 platform install ve fresh tenant.
- İkinci tenant izolasyonu, ikinci `up` no-op, yeni migration bir kez, out-of-order ret.
- Geçmiş `.sql` değişince CI fail. Yeni `.sql` geçişi.
- Role, ownership, grant, extension, constraint, function, trigger, RLS, tenant isolation, license offering, backup/restore ve setup idempotency.
- Baseline dosyasında psql meta komutu yoktur. Gövde tek Goose statement bloğudur.
- Çıkış: komut, exit code ve platform ile tenant katalog karşılaştırması yazılır.

## Sıra

Faz 0, Catalog şema kararı ile birlikte Faz 1'i açar. Sonrası Faz 1, Faz 2a, Faz 2b, Faz 3, Faz 4, Faz 5. Bu fazlar migration kaynakları ve build wiring üzerinde paralel yazmaz.
