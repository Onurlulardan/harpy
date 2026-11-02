# 05 — Doğrulanmış boşluklar

> **Tarihsel audit notu:** Aşağıdaki Catalog sayıları ve dosya yolları kaldırma öncesi envanterdir. `fix/migration-baseline-completion` branch'inde Catalog business module, 54 migration SQL kaynağı ve owner adapter'ları kaldırıldı. Baseline snapshot'ları tarihsel DDL içerdiği için platform/tenant cleanup migration'ları baseline sonrasında çalışır; beklenmeyen dolu tablolar veya dış bağımlılıklar varsa transaction fail-closed durur. Baseline dump'ın kendisi yeniden üretilmedi. `TestBaselineDumpCleanupPostgres18` mevcut platform ve tenant dump'ını PostgreSQL 18 üzerinde uygular ve cleanup sonrası Catalog önek kümesinin boş kaldığını doğrular.


Bu belge 2 Kasım 2026 çalışma ağacında doğrulanan bulguların sahibidir. Agent iş emirleri [03](03-agent-work-orders.md) içindedir. Buradaki maddeler tahmin değildir.

## Envanter

Migration dizinlerinde 2.000 `.fresh.sql` var. `contracts/runtime/v1/fresh-migration-manifest.json` 1.982 entry seçiyor: 969 platform, 1.013 tenant. Sürümler benzersiz. Hepsi `phase=expand` ve `transaction_mode=transactional`.

Owner dağılımı: iam 727, billing 208, analytics 196, platform_async 178, tenancy 131, platform_cutover 119, platform_migrator 97, documents 54, catalog 53, platformops 48, system_resource_layout 47, notifications 34, work 32, platformauthority 22, usage 15, support 13, audit 7, platform_backupreceipt 1.

`core/internal/platform/migration/release/` altında 109 `*.migration-manifest.json` var. `core/internal/platform/migration/release-index.json` duruyor. Goose sürümü `github.com/pressly/goose/v3 v3.27.2`. Repo genelindeki diğer `.sql` dosyaları (sorgu, fixture) bu 2.000 sayısının dışındadır.

## Seçilmeyen 18 SQL

`internal/migrationbundle/select.go` aynı sürümün terminal olmayan kopyasını eler. Bu 18 dosya manifestte yoktur. Baseline bunların birleşiminden çıkmaz; kaynak, seçilmiş zincirin temiz PostgreSQL'de ürettiği son şemadır.

- `core/internal/modules/analytics/migrations/tenant/fresh/20260802010000_add_analytics_dashboard_export_quota_v1.fresh.sql`
- `core/internal/modules/catalog/migrations/tenant/fresh/20260728105100_retract_catalog_empty_menu_from_v1_target_v1.fresh.sql`
- `core/internal/modules/iam/migrations/tenant/fresh/20260729100025_create_iam_installation_tenant_capability_entrypoints_v1.fresh.sql`
- `core/internal/modules/iam/migrations/tenant/fresh/20260802234002_grant_iam_support_bug_report_authority_v1.fresh.sql`
- `core/internal/modules/system_resource_layout/migrations/tenant/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql`
- `core/internal/modules/system_resource_layout/migrations/tenant/supersessions/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql`
- `core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v2/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql`
- `core/internal/modules/tenancy/migrations/platform/fresh/20260727090000_allow_tenant_executor_prepolicy_heartbeat.fresh.sql`
- `core/internal/modules/tenancy/migrations/platform/fresh/20260727090100_complete_environment_initialize_migration_state.fresh.sql`
- `core/internal/platform/async/migrations/fresh/20260718020142_allow_iam_export_reconcile_expired_attempt_v1.fresh.sql`
- `core/internal/platform/async/migrations/fresh/20260718020143_allow_iam_export_reconcile_expired_job_v1.fresh.sql`
- `core/internal/platform/async/migrations/fresh/20260802020025_allow_analytics_dashboard_export_current_unknown_v1.fresh.sql`
- `core/internal/platform/async/migrations/fresh/20260802020026_allow_analytics_dashboard_export_current_unknown_job_v1.fresh.sql`
- `core/internal/platform/async/migrations/fresh/20260802020027_allow_analytics_dashboard_export_maintenance_unknown_attempt_v1.fresh.sql`
- `core/internal/platform/async/migrations/fresh/20260802020028_allow_analytics_dashboard_export_maintenance_unknown_job_v1.fresh.sql`
- `core/internal/platform/async/migrations/fresh/20260802020036_allow_analytics_dashboard_export_maintenance_resolution_attempt_v1.fresh.sql`
- `core/internal/platform/async/migrations/fresh/20260802020037_allow_analytics_dashboard_export_maintenance_resolution_job_v1.fresh.sql`
- `core/internal/platform/cutover/migrations/fresh/20260728110020_assert_runtime_snapshot_authorities_v1.fresh.sql`

Supersession katalogları yedi dosyadadır:

- `core/internal/modules/system_resource_layout/migrations/tenant/migration-supersessions.json`
- `core/internal/modules/analytics/migrations/tenant/migration-supersessions.json`
- `core/internal/platform/cutover/migration-supersessions.json`
- `core/internal/modules/iam/migrations/tenant/migration-supersessions.json`
- `core/internal/platform/async/migration-supersessions.json`
- `core/internal/modules/catalog/migrations/tenant/migration-supersessions.json`
- `core/internal/modules/tenancy/migrations/platform/migration-supersessions.json`

## History hatası ve run lifecycle

`core/internal/platform/migrator/classify_history.go` içindeki `verifyHistoryRow` record kind, applied, owner, scope, origin, target database generation ve next minimum runtime generation karşılaştırır. Path ve SHA uyuşmazsa supersession arar. İkisi de tutmazsa `history version %d differs from verified metadata` üretir. Mesaj hangi alanın ayrıldığını yazmaz.

`Runner.Up` history karşılaştırmasından ayrı bir operasyonel akış çalıştırır: platform lock, `recoverFinalizedRelease`, `bootstrapRunLifecycle`, `ensureRunRunning`, `finalizeSuccess` ve `finalizeFailure`. Bunlar `core/internal/platform/migrator/up.go` ve `recover_finalized_release.go` içindedir. SHA/owner/scope katmanı kalkarken bu akış, Faz 0 onu legacy release generation olarak ayırmadan silinmez. Backup veya installation hâlâ bu duruma bağlıysa durum `platform.platform_migration_history` dışına taşınır. `run_id` ve generation Goose version tablosuna eklenmez.

## Goose provider ve startup

`build_goose_provider.go` ve `build_tenant_goose_provider.go` `goose.DialectCustom` ile custom `WithStore` kullanır. İkisi de seçilmiş entry sayısını Goose source listesiyle karşılaştırır.

`core/internal/composition/operator/startup.go` şema kontratı için `FreshMigrationDocuments()` çağırır ve 1.982 entry'lik corpus'u yükler. Bu çağrı migration uygulamaz. Operator initialize, migration corpus'unu parse etmeden schema contract okur.

## Skill çelişkisi

`skills/data-tenancy/SKILL.md` her migration'a checksum ve owner yazılmasını, tenant checkpoint'i tutulmasını ve en az bir eski snapshot upgrade testini ister. Skill'in işaret ettiği `docs/adr/0006-tenancy-migrations.md` bu ağaçta yoktur. Faz 4 skill metnini bu belge setine bağlar. Eksik ADR dosyası yazılmaz.

## Catalog kapısı

Catalog'un kendi migration ağacında 54 SQL vardır: 45 tenant, 9 platform. Manifestte catalog owner'lı 53 entry vardır. IAM ve Analytics tarafındaki catalog tabloları başka owner migration'larındadır. Baseline bugünkü `main` şemasından dondurulursa bu nesneler starter şemasına girer. Fresh install'da kalmayacaklarsa dump, bu nesneler kaynaktan çıktıktan sonra alınır. Ayrıntılı sıra [README](README.md) Catalog bölümündedir.

## Faz 2 sınırının açık bıraktığı testler

`platform.platform_migration_history` ve 1.982 entry assert'leri migrator paketinin dışında da durur. Örnekler: `core/internal/platform/migrator/adapters/postgres/history_store_assertions_integration_test.go`, `history_store_denials_integration_test.go`, `tenant_finalized_release_seed_integration_test.go` ve release frontier migration testleri. Faz 0, `rg` ile tam dosya listesini çıkarır. Faz 2a migrator ve composition ağacını değiştirir. Catalog removal bu branch'te ayrı ama aynı migration baseline iş akışına bağlı workstream olarak yürütülür; generic migration catalog ve IAM permission/resource catalog altyapısı korunur.
