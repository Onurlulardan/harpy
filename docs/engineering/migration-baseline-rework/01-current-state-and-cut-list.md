# 01 — Mevcut mimari envanteri ve cerrahi kaldırma haritası

## Kanıtlanmış mevcut yapı

| Alan | Repo'daki yer | Mevcut durum / karar |
|---|---|---|
| SQL kaynakları | `core/internal/modules/*/migrations/**`, `core/internal/platform/{async,cutover,migrator}/migrations/**` | 2.000 SQL dosyası; manifestte 1.982 seçilmiş entry. Yeni baseline doğrulandıktan sonra eski zincir çalışma ağacından kaldırılacak. |
| Corpus seçimi | `internal/migrationbundle/collect.go`, `corpus_sets.go`, `manifest.go`, `select.go` ve alt paketleri | Path/glob/supersession ile dosya seçip manifest üretiyor. Goose'un normal `fs.FS` migration discovery'si hedef olduğunda migration-specific seçim katmanı gereksiz. |
| Manifest writer | `internal/migrationbundle/cmd/checkfresh/**` | SQL corpus'u dolaşıp `fresh-migration-manifest.json` ve `binary-schema-contract.json` digest bağını güncelliyor. Yeni mimaride migration başına manifest/hash yok. |
| Historical release arşivi | `core/internal/platform/migration/release/**`, `release-index.json`, `load_release_registry.go`, `release_source.go`, `release_registry.go`, `release_lazy.go` | 109 historical manifest + archive/index authorization. Eski release geçmişi yeni starter'ın çalışma zamanı ihtiyacı olmayacak; bağımlılıklar temizlendikten sonra kaldırılacak. |
| Custom Goose entegrasyonu | `core/internal/platform/migrator/build_goose_provider.go`, `build_tenant_goose_provider.go` | Goose `DialectCustom` + özel `WithStore` kullanıyor. Native PostgreSQL store'a geçir; yalnız teknik olarak gereken scope/table/lock adaptörünü tut. |
| Custom history store | `core/internal/platform/migrator/adapters/postgres/**`, `Store`, `TenantStore`, `platform.platform_migration_history` | Path, SHA, owner, scope, origin, generation, run ID ve anchor kaydı tutan ayrı history sözleşmesi var. Bunları Goose'un version history işleviyle karıştırma. Migration-run/backup işlevi olan tabloları kullanımları incelenmeden topluca silme. |
| History supersession | `core/internal/platform/migration/history_supersession.go`, `core/internal/platform/migrator/classify_history.go`, `migration-supersessions.json` dosyaları | Eski dosya kimliklerini yeni dosyaya bağlayan uyumluluk katmanı. Eski zincir kaldırıldığında yeni starter'da tutulmayacak. |
| Duplicate artifact builder | `cmd/setup/release_schema_fresh.go`, `cmd/setup/release_schema_manifests.go`, `tools/internal/migrationmanifest/**`, `tools/localrelease/foundation_contract.go` | Aynı kavram farklı build/setup/test yollarına dağılmış. Tek baseline kaynağına geç; `foundation_contract.go` içindeki hard-coded digest/138-entry/frontier kontrollerini taşımak yerine foundation/full split'ini kaldır. |
| Foundation/full release orchestration | `ops/local/Dockerfile`, `ops/production/Dockerfile`, `ops/local/compose.tools.yml`, `.taskfiles/local.yml`, `.taskfiles/local-operations.yml`, `tools/localrelease/**` | İki migration artifact türü ve ayrı build/prepare aşamaları var. Tek baseline release akışına indir. `tools/localrelease` klasörünün tamamını körlemesine silme; license, identity, secret, role, extension ve bootstrap görevleri migration'dan bağımsız olabilir. ForgeCRUD Catalog package-set/release entegrasyonu ayrı [Catalog removal task](../../tasks/catalog-removal/README.md) kapsamında çıkarılır. |
| Runtime schema contract | `contracts/runtime/v1/binary-schema-contract.json`, `core/internal/platform/schemacontract/**`, `core/internal/platform/binarycontract/**` | Migration manifest digest'ine bağımlılığı kaldırılacak. API/runtime gerçekten DB schema compatibility version'ı kullanıyorsa küçük bir version contract olarak tut; 1.982 migration entry'lik catalog'a bağlama. |
| Goose bağımlılığı | `go.mod`, `core/internal/platform/migrator/build_goose_provider.go` | Goose zaten kullanılıyor. Yeni motor kurma. Normal provider, native history store ve gerekli locking ile sadeleştir. |

## Logdaki hatanın gerçek anlamı

`core/internal/platform/migrator/classify_history.go` içindeki `verifyHistoryRow` şu metadata alanlarını karşılaştırıyor: record kind, applied durumu, owner, scope, origin, target database generation ve next minimum runtime generation. Path/SHA aynı değilse ayrıca supersession kaydı arıyor. Ekteki `history version 20260722050009 differs from verified metadata` hatası bu kontrollerin en az birinin başarısız olduğunu bildiriyor; mesaj hangi alanın farklı olduğunu açıklamıyor.

Bu hata mevcut DB history'sinin mevcut release metadata'sıyla eşleşmediğini gösterir. Yeni baseline bu legacy history formatını taşımamalı. Önce hatayı geçici temiz PostgreSQL ortamında tekrarla; temiz ortamda eski migration zinciri son şemaya ulaşmıyorsa yalnız baseline'ı çıkartmak için gerekli blocker'ı teşhis et. Kullanıcının var olan Docker volume'larını izinsiz silme.

## Kesin kaldırma sırası

1. Güncel son şemayı temiz ve geçici PostgreSQL'de üret, platform ve tenant şemalarını export et, baseline doğrulamasını tamamla.
2. Yeni Goose native store/provider ile baseline ve sonrası migration testlerini çalıştır.
3. Tüm çağrı yerlerini güncelle: `rg -n "migrationbundle|FreshMigrationDocuments|LoadStartup|LoadBoundStartup|LoadReleaseRegistry|MigrationManifest|migration-supersessions|foundation-contract|foundation-applied|migration-release\.zip|platform_migration_history" .`.
4. CI, Taskfile, Dockerfile, localrelease ve Compose wiring'ini tek baseline akışına taşı.
5. Eski 2.000 SQL migration kaynağını, supersession JSON'larını, 109 release manifestini/archive/index'i ve yalnız bu sisteme ait generator/testleri sil.
6. Unused import/package, `go:embed`, golden file ve fileguard/archguard allowlist'lerini temizle. Yeni bir manifest framework'ü kurma.
7. `go mod tidy`, `go test ./...`, FileGuard, ArchGuard ve fresh-install E2E.

## Körlemesine silinmeyecekler

- `core/internal/platform/migrator/**` altındaki her şey: burada tenant executor, advisory lock, idempotent run lifecycle veya gerekli database permission işleri olabilir. Migration framework'e özel kod ile ürünün operasyonel ihtiyaçlarını çağrı grafiğiyle ayır.
- `tools/localrelease/**` klasörünün tamamı: release arşivinden bağımsız license/installation/role/secret işlevlerini koru veya doğru yere taşı.
- `contracts/runtime/v1/binary-schema-contract.json` ve `schemacontract`: bunların migration digest bağlantısı çıkarılabilir; uygulamanın gerçek minimum DB version / compatibility kontrolü varsa onun yerine küçük, statik bir değer kalabilir.
- extension kurulumu, ownership, role grants, RLS policies, function `search_path` ayarları, default privileges, tenant schema isolation ve on-premise lisans/bootstrap verileri.
- Dosya isimlerinde migration geçmese bile yedekleme/restore/identity/recovery için kullanılan tablolar ve SQL.

## Sonuçta repo'da kalması gerekenler

- Güncel state'ten türetilmiş baseline SQL artifact'ları (platform/tenant ayrımı lifecycle tarafından gerektiriliyorsa iki kapsam).
- Bundan sonra eklenecek standart Goose SQL migration'ları.
- Goose'u çağıran ince platform/tenant runner ve gerekliyse advisory lock.
- Baseline'ı ve future migrations'ı doğrulayan integration testleri; geçmiş migration dosyalarının değiştirilmesini engelleyen basit CI kuralı.
- İş migration'ından bağımsız setup, license, credentials, extensions, grants ve tenant provisioning.
