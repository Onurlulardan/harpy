# 04 — Kabul kriterleri, regresyon ve local recovery

## Definition of Done

Bu çalışma ancak aşağıdakilerin tamamı sağlandığında biter:

- [ ] Yeni kurulum 1.982 eski migration'ı sırayla çalıştırmıyor; current schema baseline'ını ve sonradan eklenmiş gerçek yeni migration'ları uyguluyor.
- [ ] Platform DB baseline'ı ve her yeni tenant schema baseline'ı gerekli lifecycle'larda uygulanıyor.
- [ ] Baseline ile oluşturulan boş DB ve eski migration zinciriyle hazırlanmış referans DB'nin ilgili catalog envanterleri eşleşiyor.
- [ ] Tüm tablolar, types, indexes, constraints, views, sequences, routines, triggers, RLS/policies, extensions, owner/grants/default privileges ve gereken statik seed data korunuyor.
- [ ] Runtime installation/environment/tenant identities, admin credentials, license/signing/recovery secrets baseline içine gömülmemiş; setup/operator kendisi oluşturuyor.
- [ ] Goose native version history tek authoritative migration history. İkinci `up` no-op, yeni migration bir kez uygulanır; out-of-order migration reddedilir.
- [ ] `platform.platform_migration_history`, supersession metadata ve custom migration manifest/release registry sadece migration tracking amacıyla artık kullanılmıyor.
- [ ] `foundation` ve `full` diye iki migration release artifact'ı yok; tek initial baseline/release akışı var.
- [ ] API/operator yalnız schema contract okumak için bütün migration corpus'unu parse etmiyor.
- [ ] Local clean bootstrap sonrası operator initialize başarılı, billing installation offering tam bir kayıt döndürüyor ve API readiness yeşil.
- [ ] Eski migrations, manifest hash, archive, index ve generator'lar current tree'den kaldırılmış; `rg` ile kalan reference'lar gözden geçirilmiş.
- [ ] `go test ./...`, `go mod tidy`, FileGuard, ArchGuard, image build ve PostgreSQL 18 integration/E2E testleri geçiyor.

## Zorunlu test matrisı

| Senaryo | Beklenen |
|---|---|
| Temiz PostgreSQL | Baseline platform schema'yı kurar; operator kurulumunu tamamlar. |
| İlk tenant | Tenant baseline yalnız hedef tenant schema'sına uygulanır. |
| İkinci tenant | İlk tenant verisi/versiyon tablosu etkilenmez; aynı tenant baseline yeni schema'ya uygulanır. |
| İkinci `goose up` | Hiçbir migration yeniden uygulanmaz. |
| Yeni SQL migration | Goose pending version'ı uygular ve history'ye yazar. |
| Out-of-order migration | CI/provider reddeder; `allowMissing` açılmaz. |
| Geçmiş migration dosyasının değiştirilmesi | CI fail eder; düzeltme için yeni migration gerekir. |
| İki migrator aynı anda | Aynı kapsam için duplicate DDL uygulanmaz; lock sonucu deterministik olur. |
| Catalog parity | Baseline source/target objects, signatures, constraints, indexes, RLS, ACL ve seed karşılaştırması geçer. |
| On-premise installation | Geçerli installation manifest ve local license/offering setup/operator yoluyla oluşur; SaaS offering fallback yoktur. |
| Local stale volume | Eski custom history formatı fresh starter ile uyumlu varsayılmaz; reset açıkça ve kontrollü yapılır. |
| Secret redaction | Connection string/parola/token/key loglarda görünmez. |

## Local hata ve mevcut volume

Ekteki log `core-migrator deploy up` çağrısında şu hatayla bitiyor:

`history version 20260722050009 differs from verified metadata`

Bu mesaj eski custom history satırının beklenen migration metadata'sıyla uyuşmadığını gösterir; hangi alanın uyuşmadığını mevcut mesaj söylemez. Yeni baseline/Goose history formatına geçildikten sonra legacy DB volume'u otomatik adapte etme. Testleri yeni, disposable PostgreSQL instance'ında çalıştır. Kullanıcının var olan Docker volume'larını otomatik silme veya `task local:destroy` çağırma; local reset adımını migration kodundan ayrı ve açık operatör işlemi olarak tut.

## Önerilen doğrulama sırası

Komut adları son wiring'e göre güncellenecek; aşağıdaki sıra değişmesin:

1. `go test ./...`
2. `go run ./tools/fileguard`
3. `go run ./tools/archguard`
4. `go mod tidy` sonrası `git diff --exit-code -- go.mod go.sum`
5. Migration dosyalarının immutable geçmiş kuralını test et.
6. Clean PostgreSQL 18 üzerinde baseline install, tenant creation ve uygulama readiness E2E.
7. İkinci Goose `up`, yeni migration apply ve concurrent migration lock integration testleri.
8. Local ve production Docker image build; tek migration baseline artifact'ının kullanıldığını doğrula.
9. `rg -n "migrationbundle|FreshMigrationDocuments|LoadReleaseRegistry|MigrationManifest|migration-supersessions|foundation-contract|foundation-applied|migration-release\\.zip|platform_migration_history" .` taramasını gözden geçir. Kalan her hit ya kaldırılmalı ya da migration framework dışında gerekli olduğu açıklanmalı.

## Agent rapor şablonu

Her faz sonunda:

- Değişen/silinen dosyalar ve her birinin sebebi.
- Çalıştırılan test/build komutları, exit code ve sonuçlar.
- Baseline catalog comparison özeti (platform + tenant ayrı).
- Korunan security/operations invariants: roles, grants, extensions, RLS, license, installation, tenant isolation, backup/restore.
- Kalan her eski migration-system reference ve neden kaldığı.
- Destructive local action uygulanıp uygulanmadığı (varsayılan: hayır).
