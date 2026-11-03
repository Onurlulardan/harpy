# Migration runtime cleanup — uçtan uca audit ve görev planı

**Durum:** P0 teşhis ve completion sözleşmesi uygulandı (`20261103000006`). Legacy migrator framework sökümü (HRY-MIG-004..009) açık. Bu doküman tüm kabul matrisinin kapandığını iddia etmez.  
**Audit edilen ref:** `main` — `8d72c6c867225d26e81c65c764473eddb0cdce63` (3 Kasım 2026).  
**Tetikleyen olay:** `task dev` akışının migration `up` ve `verify` aşamalarını geçtikten sonra `core-operator initialize` aşamasında çıkış kodu 5 ile durması.

## 1. Amaç

Harpy'nin başlangıç/tenant schema lifecycle'ını tek ve anlaşılır Goose tabanlı migration modeline indirmek; buna rağmen kurulum, tenant provisioning, role/grant, extension, RLS, license, installation, backup/restore ve recovery işlevlerini korumak.

Bu çalışma “repo içindeki migration kelimesi geçen her şeyi sil” işi değildir. Üç farklı kavram ayrılmalıdır:

1. **Schema migration engine:** Platform ve tenant PostgreSQL schema'larına Goose SQL uygulayan mekanizma. Hedefte kalır; Goose native version history tek migration truth olmalıdır.
2. **Tenant provisioning orchestration:** Yeni tenant için job/attempt/lease, schema creation, failure/retry ve activation sırası. İşlevsel olarak gerekiyorsa kalır; fakat eski migration release/manifest/checkpoint modeline bağlanmamalıdır.
3. **Legacy migration framework:** Seçilmiş migration corpus'u, her SQL için owner/scope/path/SHA/origin metadata, supersession zinciri, custom Goose Store/history, migration release generation, migration run/evidence ve manifest digest'in runtime contract'larına yayılması. Hedef mimaride kaldırılmalı veya migration dışı operasyonel gereksinimler ayrı ve küçük bir sözleşmeye taşınmalıdır.

Ana karar: **Goose migration version history schema değişikliklerinin tek kaynağı olsun.** Tenant job state'i provisioning sürecini anlatsın; migration dosyalarının ikinci bir versiyon/checkpoint defterine dönüşmesin.

## 2. Son hata: kanıtlanan akış ve henüz kanıtlanmayan nokta

Gönderilen logdaki sıra:

| Aşama | Sonuç |
|---|---|
| Docker image build | Başarılı |
| Local PostgreSQL / RabbitMQ / MinIO / KMS / SMTP / ClamAV | Healthy |
| Local release prepare | Başarılı |
| `core-migrator deploy up` | Başarılı; `20261103000005` uygulandı |
| Runtime grants | Başarılı |
| `core-migrator deploy verify` | Başarılı; database version `20261103000005` |
| Tenant migration executor | Healthy |
| Installation delivery prepare | Başarılı |
| `core-operator` | **Başarısız:** `environment initialize is resumable`, stage `initialize`, exit code 5 |

Dolayısıyla log, platform migration apply/verify hatası göstermiyor. İlk gözlenen hata tenant provisioning/onboarding'i tamamlaması gereken operator katmanında.

### Kod üzerinden izlenen yol

- `core/internal/composition/operator/startup.go` — `initializeEnvironment()`, `coordinator.Initialize(...)` çağrısını yapıyor; hata yalnızca `initialize error: %v` olarak stderr'e yazılıyor.
- `core/internal/composition/operator/process_result.go` — `ErrFailed` ve `ErrFenced` terminal sayılıyor; diğer hatalar `OperatorResumable` (logdaki exit code 5) oluyor.
- `core/internal/orchestration/environmentoperator/initialize.go` — deployment policy, plan hazırlığı, onboarding accept ve devam akışını yürütüyor.
- `core/internal/orchestration/environmentoperator/continue.go` — accepted/migrating state'lerinde tenant executor readiness bekliyor, claim alıyor, migration'ı bekliyor; sonra finalization'a geçiyor.
- `core/internal/orchestration/environmentoperator/migrate.go` — job'ı enqueue edip poll ediyor. Job `succeeded` olduğunda `refreshMigration()` çağrılıyor.
- Aynı dosyadaki `refreshMigration()`, onboarding `Accept(plan)` sonucunun beklenen `migrationNextState` değerine gelmesini ve revision'ın claim revision'ından büyük olmasını istiyor. Accept hata verirse veya state/revision koşulu sağlanmazsa yalnızca `ErrResumable` dönüyor.

**Önemli teşhis sınırı:** Log, `ErrResumable`'ın hangi çağrı noktasından geldiğini göstermiyor. `refreshMigration()` güçlü bir şüpheli çünkü tenant job'ın başarı durumundan onboarding state machine'e geçiş yaptığı kritik sınır burası; ancak yalnızca bu logla kesin kök neden denemez. `initialize.go`, `continue.go` ve `migrate.go` içindeki diğer resumable dönüşleri de aynı genel mesajı üretebilir. İlk task tahminle SQL değiştirmek değil; hatanın tam stage/cause/state/revision bilgisini yakalamak ve aynı failure'ı disposable DB üzerinde tekrar üretmektir.

### Doğrulanmış sözleşme uyuşmazlığı / riskler

- Tenant executor adapter'ı `core/internal/platform/migrator/tenantmigrationexecutor/adapters/postgres/sql/platform/complete.sql` üzerinden `platform.tenancy_tenant_migration_complete_v3(...)` çağırıyor.
- `core/internal/platform/baseline/sql/20261103000004_platform_tenant_executor_completion_baseline.fresh.sql` içinde tanımlı completion fonksiyonu `platform.tenancy_tenant_migration_complete_v2(...)`. Bu dosyada `complete_v3` tanımı yok.
- `contracts/storage/owners/tenancy.json` hem `complete_v2` hem `complete_v3` routine'lerini tanımlıyor. Gerçek DB'de hangi imzanın mevcut/grant'li olduğu ve 3 Kasım baseline'ının gerçekten hangi routine'leri içerdiği kontrol edilmelidir.
- Aynı `20261103000004` completion fonksiyonu tenant Goose history'sinde iki sabit versiyonun `20261028000001` ve `20261102000001` uygulanmış olmasını bekliyor; kontrol başarısızsa `tenant migration checkpoints are incomplete` üretiyor. Bu, eski manifest/checkpoint varsayımının yeni Goose modeline taşınmış olduğunu gösteriyor.
- `contracts/runtime/v1/binary_schema_contract.go` hâlâ `MigrationManifestSHA256` içeriyor. `core/internal/platform/postgres/probe_schema_contract.go` bunu schema readiness check'e geçiriyor.
- `core/internal/platform/postgres/foundation_baseline.go` Goose version kayıtlarında iki sabit platform version'ı sayıyor ve eski binary migration-manifest digest alanlarını foundation/readiness sonucuna eşliyor.
- Platform migrator `Up`/`Verify` akışları hâlâ `selected` migration metadata, `classifyHistory`, custom run lifecycle, finalization ve stored evidence kullanıyor. Bu, Goose native history tek otorite olma hedefiyle çelişen aktif runtime bağımlılığıdır.
- Tenant executor readiness/job/claim sözleşmeleri `release_id`, `runtime_generation`, `database_generation`, `migration_manifest_sha256` taşıyor. Bu alanların gerçek runtime compatibility ihtiyacı mı, eski release framework'ünden kalma metadata mı olduğu tek tek gerekçelendirilmelidir.

Bu bulgular, adapter/fresh baseline SQL, owner aggregate ve DB fonksiyon kataloglarının tek bir migration cutover'ı olarak ele alınmasını gerektirir.

## 3. Hedef sınırlar

### Korunacaklar

- Goose ve standart timestamp'li SQL migration'ları.
- Platform DB ve her tenant schema'sı için doğru scope'ta migration uygulama.
- Goose native PostgreSQL version store; platform ve tenant schema isolation.
- Aynı schema için eşzamanlı migrator'ları serialize eden gerekli lock.
- Tenant provisioning job/attempt/lease/retry ve activation orkestrasyonu (consumer/call graph bunu doğruladığı sürece).
- Installation manifest, initial admin, credential delivery, local license/offering, role/grant, extension, RLS, schema ownership, backup/restore, recovery ve secret projection.
- `contracts/storage/owners` gibi genel DB ownership/security governance; owner manifestleri generic kaldığı sürece.

### Kaldırılacak / sadeleştirilecekler

- Migration başına path/SHA/owner/scope/origin/phase/transaction metadata taşıyan eski runtime selected-manifest framework'ü.
- Migration-specific supersession alias zinciri ve historical release manifest/archive/index.
- Goose native store'un zaten tuttuğu version/applied bilgisini ikinci kez tutan custom history/Store.
- Sadece eski migration release modelinin çalışması için bulunan release generation, migration run/evidence, bootstrap receipt, schema contract ve generation state'i. Önce backup/restore, installation ve operational consumers taranmalı; gerçek operasyonel state gerekirse migration framework dışına taşınmalı.
- Goose history yanında tutulan sabit version listeleri ve migration checkpoint/requirement catalog'ları.
- Runtime binary schema contract içindeki `migration_manifest_sha256` bağı; gerekli bir minimum schema compatibility değeri varsa bunu küçük ve migration-digest'ten bağımsız bir sözleşmeye dönüştür.
- Docker image'a migration manifest kopyalama, setup/localrelease manifest argümanı, Taskfile manifest verify ve gereksiz custom migrator command wiring.
- Eski routine sürümleri ve owner manifest kayıtları; yalnızca consumer yokluğu ve forward migration planı kanıtlandıktan sonra.

## 4. Güvenli uygulama ilkeleri

1. **Önce loglama ve disposable reproduction; sonra SQL düzeltme.** Genel `ErrResumable` mesajını gerçek cause'u kaybetmeden sar.
2. **Applied Goose migration dosyalarını değiştirme.** `20261103000003`, `00004`, `00005` gibi uygulanmış versiyonların içeriği immutable kabul edilir. Düzeltme yeni timestamp'li forward migration ile yapılır. Fresh baseline source değişikliği ayrıca yeni kurulum artifact'ında doğrulanır.
3. Kullanıcının mevcut local named volume'larını, PostgreSQL datadir'ini veya production DB'yi silme/resetleme. Her repro disposable Compose project/container/volume kullanır.
4. `core/internal/platform/migrator` ve `tools/localrelease` klasörlerini topluca silme. Her dosya için consumer/call graph ve security/ops invariant raporu gerekir.
5. Yeni custom registry, manifest digest zinciri, release archive, supersession veya özel Goose Store başka isimle tekrar oluşturulmaz.
6. Platform schema migration ile tenant provisioning state machine aynı şey değildir. Birini sadeleştirirken diğerinin durable retry/lease/activation garantileri kaybedilmez.
7. DB routine, grants, owner aggregate, embedded SQL ve baseline artifact aynı değişiklik setinde tutarlı güncellenir. Tek integration owner shared aggregate dosyalarına yazar.
8. Hataları testten kaçırmak için assertion gevşetilmez; integration test atlanmaz; migration geçmişi yeniden yazılmaz.

## 5. İş sırası

| Sıra | Task | Öncelik | Çıktı |
|---|---|---:|---|
| 0 | Runtime failure instrumentation + disposable reproduce | P0 | Hatanın gerçek çağrı noktası, SQLSTATE, job/operation state ve revision farkı |
| 1 | Tenant completion / onboarding contract parity | P0 | Executor adapter, Goose SQL, tenancy owner routine ve onboarding state transition aynı kontrata bağlanır |
| 2 | Full `task dev` initialize E2E + retry safety | P0 | Fresh install ve interrupted/resume senaryoları geçer |
| 3 | Legacy migration runtime inventory + target contract | P1 | Keep/remove/move matrisi, tek Goose history kararı ve referans envanteri |
| 4 | Manifest/history/release framework extraction | P1 | Custom migrator metadata, release evidence ve digest bağı kaldırılır |
| 5 | Tenant executor decoupling | P1 | Provisioning job'ları sürer; release/manifest/checkpoint framework bağı kalkar |
| 6 | Setup / Docker / Taskfile / contracts cleanup | P1 | Runtime image ve task akışı tek Goose kaynaklı sade yol olur |
| 7 | Forward migration + owner/permission/fixture alignment | P1 | DB routine sürümleri ve owner manifests tutarlı, deployed DB forward-only |
| 8 | Final audit, docs and CI gates | P2 | Eski runtime referansları yok/istisnaları gerekçeli; E2E ve güvenlik kapıları yeşil |

Detaylı agent iş emirleri [01-agent-work-orders.md](01-agent-work-orders.md), kabul ve recovery [02-acceptance-and-recovery.md](02-acceptance-and-recovery.md) içindedir.

## 6. Bitti sayılma şartı

- Son logdaki `task dev` zinciri disposable PostgreSQL üzerinde baştan sona başarılıdır.
- Tenant migration job succeeded olduğunda onboarding operation doğru state'e ve daha yüksek revision'a deterministik geçer; operator finalization/credential delivery/authority publication tamamlanır.
- Operator hatası çıktığında stage + safe cause + SQLSTATE + relevant state/revision görünür; generic resumable tek başına bırakılmaz.
- Completion adapter çağırdığı routine imzası, baseline/forward migration tanımı, storage owner manifesti ve role grants ile birebir eşleşir.
- Goose migration history platform ve tenant schema için tek authoritative applied-version kaynağıdır.
- Runtime contract'larda migration manifest digest/path/SHA/supersession/custom version ledger bağı kalmamıştır; gerçek runtime compatibility ihtiyacı varsa küçük ve açık bir schema compatibility contract vardır.
- Tenant executor durable job/lease/retry güvenliği korunur fakat eski release-generation/manifest/checkpoint state'e bağımlı değildir.
- Applied migration'lar immutable; local/production volume resetlenmemiştir.
- `task dev`, clean install, second run/no-op, tenant retry/recovery ve API/worker readiness testleri geçer.

## 7. Uygulama notu — 3 Kasım 2026

P0'ın kodla kapanan kısmı:

- Operator `ErrResumable` dönüşleri stage, error code, redacted cause ve varsa SQLSTATE taşır. JSON failure kaydı `error_code`, `detail` ve `sqlstate` alanlarını geriye uyumlu ekler.
- `platform.tenancy_tenant_migration_complete_v3` yeni forward migration `20261103000006` ile değiştirildi. Applied `20261103000003`–`00005` dosyalarına dokunulmadı.
- Completion, tenant şemasındaki Goose history için tek minimum sürüm (`20261102000001`) kullanır. İki sabit checkpoint sürümünü saymaz. Job `succeeded` ile onboarding `migration_ready` / `authority_pending` geçişi aynı routine içindedir ve kayıp cevaptan sonra tekrar çağrı aynı final state'i korur.

Açık kalan iş: HRY-MIG-003 disposable `task dev` E2E, HRY-MIG-004..009 legacy manifest/release framework sökümü ve CI fresh-install kapısı. Bu ortamda Docker olmadığı için tam `task dev` doğrulanmadı.
