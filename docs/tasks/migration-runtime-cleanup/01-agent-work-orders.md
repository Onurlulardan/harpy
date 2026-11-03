# 01 — Agent iş emirleri ve dosya bazlı çalışma planı

Bu görevler tek integration branch üzerinde ve sırayla yürütülür. P0 tamamlanmadan migration framework sökümü başlamaz. Shared contract/storage owner/baseline aggregate dosyaları tek entegrasyon sorumlusunun yazma alanıdır.

## P0 — HRY-MIG-001: Operator initialize gerçek hata noktasını yakala

**Sahiplik:** `core/internal/orchestration/environmentoperator/**`, `core/internal/composition/operator/**`, ilgili tenancy PostgreSQL adapter testleri.

### Yapılacaklar

1. `environmentoperator.Initialize`, `continueOperation`, `waitClaimReady`, `claim/renew`, `migrate`, `refreshMigration`, `finalize` aşamalarının her birinde safe stage identifier üret. Hata wrap edilirken `%w` ile orijinal cause korunmalı.
2. `ErrResumable` dönüşlerinin hangi alt nedenleri örttüğünü tek tek ayır:
   - deployment policy create conflict / database error,
   - onboarding accept error,
   - tenant migration enqueue/get error,
   - job bulunamaması veya plan binding uyuşmazlığı,
   - migration deadline aşımı,
   - onboarding refresh state/revision mismatch,
   - lease acquire/renew conflict,
   - finalizer callback / receipt error.
3. `refreshMigration()` hatasında şu alanları secret içermeden logla: operation ID, tenant ID, job ID, job state, deployment mode, onboarding state before/after, claim revision, returned revision, expected next state, SQLSTATE ve wrapped cause. Lease token, DSN, parola, admin credential, signing key, raw manifest veya secret path loglama.
4. `core/internal/composition/operator/process_result.go` içindeki machine-readable failure result'a safe `error_code` / `detail` alanı eklemeyi değerlendir. Var olan stdout JSON consumer contract'larını bul; geriye uyumlu alan ekle veya stderr structured log kullan. Kullanıcıya secret sızdırma.
5. `core/internal/composition/operator/startup.go` genel `initialize error: %v` yazımı yerine stage ve error chain'i korusun. Hataları tüm sistem için generic `ErrResumable`'a çevirmeden önce sınıflandır.
6. Unit test: her resumable edge için alt hata ayrımı; refresh success/state mismatch/revision mismatch/Accept error; context timeout; lease conflict; finalizer failure.
7. En az bir disposable PostgreSQL integration testinde operator initialize'ın tam çağrı zincirini çalıştır ve başarısızlık durumunda teşhis verisinin kaybolmadığını doğrula.

### Kabul

- Aynı hata tekrar oluştuğunda hangi çağrının resumable döndürdüğü tek log/result üzerinden belirlenir.
- Cause chain ve SQLSTATE kaybolmaz.
- Secret/credential/DSN redaction testleri geçer.
- Hatanın stage'i kullanıcıya görünen `stage: initialize` bilgisinden daha ayrıntılıdır (ör. `tenant_migration.refresh_onboarding`).

## P0 — HRY-MIG-002: Tenant migration completion ile onboarding transition kontratını eşitle

**Sahiplik:** Tenant migration executor adapter, tenancy onboarding/tenantmigration adapter, tenancy-owned SQL functions, baseline/forward migration.

### İlk doğrulama (kod değiştirmeden)

1. `core/internal/platform/migrator/tenantmigrationexecutor/adapters/postgres/sql/platform/complete.sql` dosyasının `complete_v3` çağrısını gerçek main baseline ve DB fonksiyonu ile karşılaştır.
2. `core/internal/platform/baseline/sql/20261103000004_platform_tenant_executor_completion_baseline.fresh.sql` içinde `complete_v2` tanımının bulunduğunu ve `complete_v3` tanımının bulunmadığını göz önünde tut. Bunun gerçek DB'deki function list, signature, owner, SECURITY DEFINER, `search_path`, execute grants ve migration order ile tutarlı olup olmadığını disposable PostgreSQL'de sorgula.
3. `contracts/storage/owners/tenancy.json` routine list + `routine_execute_grants` alanlarını source SQL, generated grants ve runtime caller ile karşılaştır. V2/V3 function girişlerinin hangi adapter ve testler tarafından çağrıldığını çıkar.
4. Completion'ın döndürdüğü job state ile `tenancy_onboarding_accept_v2`'nin onboarding state/revision davranışını aynı transaction boundary boyunca incele. `refreshMigration()` beklenen `migrationNextState()` için gerçek transition'ın nerede üretildiğini kanıtla.
5. Tenant migration completion SQL'inin Goose history check'ini incele. Şu anki `20261103000004` iki sabit tenant version'ı (`20261028000001`, `20261102000001`) sayıyor ve eksikse `tenant migration checkpoints are incomplete` diyor. Bu sabit iki version kontrolünün hedef mimaride kalıp kalamayacağını ADR/runtime contract üzerinden karara bağla; tarihsel release checkpoint listesini uygulama SQL'inde bırakma.

### Düzeltme

6. Tek authoritative completion contract belirle. Adapter, database routine signature, result columns, tenancy owner, generated permission manifest ve integration fixture aynı routine version/shape'e bakmalı. Unused V1/V2/V3 alias'larını sonradan temizle; ancak deployed DB için applied migration dosyasını düzenleme.
7. Job completion ile onboarding operation state transition ya atomik tek DB routine olmalı ya da durable/idempotent iki aşamalı protocol açıkça tanımlanmalı. `job.state=succeeded` ile operation state `migrating` arasında kalıcı ve görünmez bir ara durum oluşmamalı.
8. Retry/idempotency kurallarını açıkla: executor complete response kaybolursa, operator `Get()` tekrar succeeded görürse, onboarding refresh tekrarlanırsa aynı final state elde edilmeli.
9. Goose version check'i migration engine'in native state'i üzerinden genel olarak doğrula; migration version'larını uygulama SQL'ine tarihsel sabitler halinde gömme. Eğer runtime için minimum schema floor gerekiyorsa migration digest yerine açık `minimum_schema_version` compatibility contract kullan.
10. Applied migration'lar için yeni forward SQL migration yaz. Fresh baseline üretim kaynağı farklı güncelleniyorsa fresh install ve mevcut DB upgrade'i ayrı doğrula.

### Kabul

- Adapter ve gerçek DB'de tek bir function signature/return shape var.
- Completion adapter'ın çağırdığı routine baseline/forward migration'da mevcut, owner'ı ve grants'i doğru.
- Tenant job succeeded ile onboarding state/revision transition tutarlı ve idempotent.
- Executor tamamlandıktan sonra operator `refreshMigration()` beklenen state ve revision'ı görür.
- İki tenant'ın schema/history state'i birbirine karışmaz.
- Sabit tarihsel checkpoint kontrolü yoktur; gerekçeli minimum schema compatibility kontrolü varsa native Goose state'ten türetilir.

## P0 — HRY-MIG-003: `task dev` local initialize E2E ve resume güvenliği

**Sahiplik:** `.taskfiles/local.yml`, local Compose, Go integration/E2E tests, CI.

1. Logdaki komut sırasını birebir disposable Compose project ve PostgreSQL 18 ile otomatikleştir: build → infra → bootstrap/up → grant/verify → tenant executor → installation delivery prepare → operator initialize → authority publication → delivery → services/readiness.
2. Birinci çalıştırmada single-tenant kurulum tamamlanmalı. Logda başarı yalnız `core-migrator deploy verify` sonrasında değil; operator initialize, tenant activation, first admin credential delivery, authority publication, API/worker readiness sonrası ilan edilmeli.
3. Aynı `task dev` ikinci kez çalıştırıldığında idempotent/no-op olmalı; duplicate tenant, user, onboarding receipt, license offering veya schema creation oluşmamalı.
4. Tenant executor çalışırken process kill/lease expiry/worker restart senaryolarını ekle. Migration rollback veya outcome ambiguous ise güvenli reconcile/cleanup path çalışmalı; completed Goose migration tekrar DDL çalıştırmamalı.
5. İki ayrı tenant oluşturup tenant schema/history isolation'ını, bir tenant'taki migration fail/retry'nin diğerini etkilemediğini doğrula.
6. Kullanıcının mevcut `harpy-local-postgres` volume'una veya başka persistent volume'a dokunmayan disposable project name üret; test cleanup sadece kendi resource ID'lerini kaldırsın.

### Kabul

- CI fresh local path'te operator initialize dahil baştan sona geçer.
- İkinci `task dev` no-op veya deterministik resume olur.
- Crash/retry ve iki tenant izolasyonu testleri geçer.
- Gerçek local volume silinmez ve test harness bunu doğrular.

## P1 — HRY-MIG-004: Legacy migrator framework envanteri ve kaldırılma matrisi

**Sahiplik:** Read-only audit; sonra tek migration integration owner.

Aşağıdaki alanların full call graph'ını ve runtime consumer'larını çıkar; her dosya/alan için KEEP / REMOVE / MOVE kararını, gerekçeyi, owner'ı, testini yaz:

### Migrator runtime ve custom history

- `core/internal/platform/migrator/**`: `selected`, `historyAliases`, `classifyHistory`, `readHistory`, `recoverFinalizedRelease`, `bootstrapRunLifecycle`, `ensureRunRunning`, `finalizeSuccess/failure`, `verifyStoredEvidence`, platform/tenant custom stores, migration-run IDs, release generation references.
- `core/internal/platform/migration/**`: historical release manifests/index, source registry, supersession maps, custom selection/manifest logic. Genel storage owner/source governance başka bir işlev sağlıyorsa onu ayrı tut.
- `core/internal/platform/migrator/adapters/postgres/**`: custom history table, SHA/path/owner/scope metadata, database generation manifest, migration run/evidence, receipt, release generation, tenant requirement/checkpoint tables ve SQL routines.
- `core/internal/platform/tenantmigrationexecutor/**` ve `core/internal/platform/migrator/tenantmigrationexecutor/**`: Goose execution vs provisioning orchestration sınırını ayır; engine adapter kalırken obsolete metadata fields çıkarılsın.
- `core/internal/platform/baseline/**`: foundation preflight'in migration-manifest hash ve sabit historical version listesi bağı.

### Runtime contract / build and setup

- `contracts/runtime/v1/fresh-migration-manifest.json`, `fresh_migration_manifest.go`, migrator result schema/Go types, `contracts/postgres/v1/**`, `binary_schema_contract.go`, `probe_schema_contract.go`.
- `Dockerfile` / local and production Dockerfile: `migration-manifest.json` copy/ownership/mode and core-migrator binary.
- `tools/localrelease/**`, `cmd/setup/**`: manifest argument and migration artifact generation vs retained roles/extensions/license/installation/bootstrap behavior.
- `.taskfiles/local.yml`, `Taskfile.yml`, CI workflows: old migration verify task, manifest prepare and duplicate `bootstrap/migrate` steps.
- `contracts/storage/owners/*.json`, `contracts/storage/migration-streams/**`, owner routines and permission generation. Generic storage governance must be classified separately from old migration runtime selection.

### Kabul

- File-by-file inventory with consumer, removal/move operation, target owner and verification test.
- Every kept migration-specific datum has a written reason beyond “currently referenced”.
- Goose native history is confirmed sufficient for platform, tenant, concurrency and recovery semantics.
- No source deletion begins from simple grep results alone.

## P1 — HRY-MIG-005: Platform migrator'ı Goose native history'ye indir

1. Pinned Goose version API'sini ve `NewProvider` / native PostgreSQL store / `WithTableName` / locker davranışını platform ve per-schema tenant lifecycle'da integration testle kanıtla.
2. `core/internal/platform/migrator` içinde migration file selection, selected source digest, supersession, custom Store/history classification ve run metadata'nın migration truth rolünü kaldır.
3. İstenen platform/tenant database/schema için tek provider çağrısı; Goose native `goose_db_version` tablo adı/schema scope davranışını açıkça seç.
4. Concurrency için tek lock strategy bırak; özel advisory lock ancak Goose locker karşılamayan somut ihtiyacı testle gösterirse kalır.
5. Goose `Up`, `UpTo`, `Status`, `Version` gibi davranışların hangi minimal surface'inin kullanılacağını belirle. `plan/verify` JSON contract'ı artık şart değilse kaldır; gerekli runtime compatibility check'i ayrı tut.
6. Applied version history'nin path/SHA/owner/scope/origin/generation metadata'sını Goose table'a custom kolon ekleyerek geri getirme.
7. Migration failure output: version + filename + scope + SQLSTATE; DSN/password/secret redacted.

### Kabul

- Goose native version history tek applied-version ledger'dır.
- Platform install, normal up, no-op second up, failed transaction retry ve concurrent execution deterministic.
- Tenant scope için schema-specific history birbirinden ayrıdır.
- Custom manifest/supersession/history selection runtime path'i kalmamıştır.

## P1 — HRY-MIG-006: Tenant executor'ı release manifest ve migration checkpoint modelinden ayır

**Kapsam:** Tenant provisioning iş kuyruğu korunur; migrator version ledger sadeleşir.

1. `TenantMigrationJob`, `TenantMigrationClaim`, `Release`, `ReadyReceipt`, completion `Receipt` ve store SQL'de `release_id`, `runtime_generation`, `database_generation`, `migration_manifest_sha256`, `origin`, `migration_run_id`, `previous_run_id` alanlarının tüketicilerini bul.
2. Her alanı deployment fencing, schema compatibility, legacy manifest veya operational receipt açısından sınıflandır; yalnız “kullanılıyor” diye tutma.
3. Executor fencing ihtiyacı gerçekse minimum generation/version/build contract gibi dar bir sözleşmeye taşı; tüm migration corpus digest'ini her job'a kopyalama.
4. Job/attempt lease, heartbeat, failure, ambiguous outcome, safe schema cleanup/retry ve tenant isolation semantiğini koru.
5. Completion result'tan `catalog_sha256` ve `requirement_count` gibi eski catalog/requirement isimlerini yalnız gerçek consumer kalmadıktan sonra kaldır; adapter, contract, owner, generated grant ve tests birlikte güncellensin.
6. Goose schema history kontrolünü tenant migration job receipt'ine ayrı migration catalog olarak ekleme.

### Kabul

- Tenant provisioning worker durdurulabilir/restart edilebilir; lease-expiry/retry ve ambiguous outcome güvenlidir.
- Migration applied versions yalnız tenant schema'sındaki Goose store'dan okunur.
- Tenant provisioning job state'i Goose history'yi kopyalamaz.
- Runtime manifest/release digest alanları kaldırılmış veya belirli bir operational threat model ile belgelenmiş az sayıda alana düşürülmüştür.

## P1 — HRY-MIG-007: Local release, Docker ve Taskfile sadeleştirmesi

1. `tools/localrelease` ve `cmd/setup` içinde migration manifest yükleme/parse/package/verify bağımlılıklarını çıkar.
2. License, installation, identity, credentials, secrets, extension install, role/grant, ownership, backup/restore ve recovery fonksiyonlarını koru.
3. Image'dan `fresh-migration-manifest.json` → `migration-manifest.json` copy'sini kaldır; gerekli migration SQL'leri Goose filesystem/embed kaynağından taşınsın.
4. `.taskfiles/local.yml` içinde `bootstrap` ve `migrate` görevlerinin aynı işi iki kez yapıp yapmadığını incele. Tek idempotent lifecycle path'i belirle; prepare/apply/grant/verify sırası açıkça yaz.
5. `Taskfile.yml` içindeki `migration:verify` komutunu yeni Goose contract'a göre yenile veya artık gereksizse sil. Eski `core/internal/platform/migration` paketini test ediyor diye migration framework'ünü gizlice koruma.
6. Local ve production Compose/Dockerfile aynı migration entrypoint/provider contract'ını kullanmalı.

### Kabul

- Fresh install path'i tek bir Goose-baseline akışından geçer.
- Manifest file ve parametreleri runtime image/CLI/Taskfile'da yoktur.
- Non-migration localrelease functionality regression olmamıştır.
- `task dev` build, bootstrap, migration, initialize ve readiness adımlarını tek sırada yürütür.

## P1 — HRY-MIG-008: Forward-only DB cleanup + owner/grant consistency

1. Deployed DB'de mevcut routine/table/function inventory al. Her routine için signature, owner, security definer, search_path, execute grant, direct caller ve version history yaz.
2. Applied migrations dosyalarına dokunmadan yeni timestamp'li forward migration ile stale routine/old custom history cleanup yap. Data inventory ve backup/recovery bağımlılığı doğrulanmadan DROP yok.
3. Fresh baseline artifact ve upgrade-from-current-DB yolu ayrı test edilir. İkisinin son schema catalog comparison sonucu aynı olmalıdır.
4. `contracts/storage/owners/tenancy.json` içindeki duplicate/unused `complete_v1/v2/v3`, `get_v1/v2`, `accept_v1/v2` gibi routine registrations için gerçek source + consumer parity sağla. Adı eski diye drop etme; consumer ve deploy support window doğrulanmalı.
5. Generated routine grants, owner manifests, schema snapshot fixtures, test expected counts ve SQL embed'lerini canonical generator ile güncelle.
6. `harpy_runtime`, `harpy_migrator`, `harpy_owner`, backup/restore ve executor minimum grants modelini regression testle doğrula.

### Kabul

- Fresh install ve mevcut supported DB forward-upgrade aynı contract setinde birleşir.
- Stale runtime function/permissions kalmaz.
- SECURITY DEFINER, search_path, owner/grants, RLS ve tenant schema boundary korunur.
- SQL cleanup idempotent değilse migration transaction/fail-closed davranışı belgelenir.

## P2 — HRY-MIG-009: Final reference audit, docs and CI

1. Aşağıdaki aramaları repo genelinde çalıştır; hit'leri KEEP / REMOVE / MOVE / TEST FIX olarak sınıflandır:
   - `fresh-migration-manifest`, `migration_manifest_sha256`, `MigrationManifestSHA256`
   - `platform_migration_history`, `classifyHistory`, `historyAliases`, `migration-supersessions`
   - `release_generations`, `schema_contracts`, `migration_runs`, `bootstrap_receipts`
   - `tenant_migration_requirement_catalog`, `tenant_migration_checkpoints`
   - `migration_run_id`, `previous_run_id`, `selected migration`, `release generation`
   - `migration-manifest.json`, `-migration-manifest`
2. Docs, README, ADR index, skills, task descriptions, CI labels ve error messages'i yeni hedef mimariyle hizala. Tarihsel audit notlarını silmek yerine tarihlerini/retrospective durumunu belirt.
3. FileGuard/ArchGuard ve custom tests'te eski migration assumptions kalmasın.
4. CI'ye disposable PostgreSQL 18 fresh install + tenant E2E ve Goose immutability checks ekle.
5. Başarı raporunda kalan her match'in neden kaldığı yazılır.

### Kabul

- Eski migration framework runtime references yoktur veya tek tek gerekçelidir.
- Migration history için ikinci truth yoktur.
- CI fresh platform + two-tenant schema isolation + retry path + operator initialization testini koşar.
- Docs ve code aynı mimariyi anlatır.
