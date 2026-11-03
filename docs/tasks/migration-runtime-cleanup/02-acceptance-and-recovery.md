# 02 — Kabul kriterleri, test matrisi ve recovery

## Definition of Done

### Runtime initialize

- [ ] Operator initialize hatasının hangi alt çağrıdan kaynaklandığı stage, wrapped cause ve SQLSTATE ile kanıtlanıyor.
- [ ] Log/result alanlarında secret, DSN, password, token, private key veya raw credential yok.
- [ ] Tenant executor job succeeded olduğunda onboarding operation expected next state'e geçiyor ve revision monoton artıyor.
- [ ] Completion SQL adapter çağrısı, DB routine signature/return type, actual applied Goose migration, storage owner registration ve grants birebir eşleşiyor.
- [ ] Completion idempotent; worker response loss ve operator retry sonrası aynı final result dönüyor.
- [ ] Finalization, tenant activation, IAM identity/credential, platform authority publication ve installation delivery complete.

### Migration engine

- [ ] Goose native PostgreSQL version store applied migration truth'un tek kaynağı.
- [ ] Platform ve tenant schema'larında history izolasyonu var; iki tenant birbirinin version state'ini görmüyor.
- [ ] Baseline yeni DB'ye bir kez uygulanıyor; ikinci up no-op.
- [ ] Yeni timestamp'li migration uygulanıyor; rollback/failed SQL ve retry davranışı testli.
- [ ] Out-of-order uygulama kapalı.
- [ ] Geçmiş applied migration dosyası değiştirilince CI fail ediyor.
- [ ] Custom selected corpus, per-migration SHA/owner/scope/origin/supersession, custom Store history veya duplicate release ledger yok.

### Tenant provisioning

- [ ] Job/attempt/lease/readiness/retry/heartbeat/ambiguous outcome state transitions testli ve durable.
- [ ] Goose version history, tenant provisioning job receipt'lerinden bağımsız.
- [ ] Eski release manifest digest/generation alanları kaldırılmış veya threat model'i açıklanmış minimum operational contract'a indirgenmiş.
- [ ] Schema cleanup yalnız doğrulanmış, aktive olmamış ve lease'i bulunmayan tenant schema'sına uygulanıyor; live tenant schema'ya dokunmuyor.
- [ ] İkinci tenant kurulumu ilk tenant'ın schema'sını/history'sini etkilemiyor.

### Setup/security/operations

- [ ] Runtime image'da migration-manifest.json yok; manifest argümanı/task artık çağrılmıyor.
- [ ] tools/localrelease license, installation, identity, secrets, extensions, owner/roles/grants, backup/restore/recovery işlevleri çalışıyor.
- [ ] Storage owner ve security grant aggregate'leri source SQL ile aynı.
- [ ] RLS/policies, SECURITY DEFINER, function search_path, extensions, privileges, schema ownership ve secret redaction regression'ları geçiyor.
- [ ] Local/production DB named volume resetlenmedi; destructive işlemler yalnız disposable environment'da yapıldı.

## Test matrisi

| Test | Beklenen |
|---|---|
| Empty PostgreSQL 18 platform DB | Baseline uygulanır, Goose history kaydı oluşur, verify/readiness geçer |
| Fresh tenant schema | Tenant baseline uygulanır, yalnız tenant schema goose_db_version etkilenir |
| Two tenants | Tenant A/B version histories ayrı; schema data izolasyonu korunur |
| Repeated task dev | İkinci çalıştırma idempotent/no-op veya güvenli resume; duplicate provisioning yok |
| Executor readiness unavailable | Operator güvenli şekilde bekler/timeout verir; gerçek cause görünür |
| complete_vN signature absent | Apply/verify veya startup açık contract error verir; resumable diye maskelenmez |
| Tenant completion success | Job succeeded, operation state/revision geçişi doğru, operator finalization devam eder |
| Completion response lost | Retry duplicate receipt/job completion yaratmaz; final state aynı |
| Onboarding accept returns wrong state | Expected/actual state and revision safe diagnostic; silent generic failure yok |
| SQL error | Goose transaction/SQLSTATE/file/version raporlanır; connection secret yok |
| Worker crash before/after DDL commit | Goose history + provisioning job state reconcile edilir; DDL duplicate değil |
| Expired lease | New executor claim allowed only when previous lease is expired; no concurrent DDL |
| Concurrent platform migration | Duplicate DDL yok; lock deterministic |
| Concurrent tenant migration | Same schema lock serialized; separate tenant can progress |
| Failed tenant migration | Transactional migration rolled back; retry works |
| Nontransactional migration (if supported) | Explicitly documented and resumable outcome; no implicit retry |
| Previous migration mutation | CI rejects changed/deleted migration source |
| Fresh install vs upgrade | Object/routine/grant/RLS catalogs equal after forward-upgrade |
| Backup/restore | Migration state, tenant data, authority and secrets restore correctly |
| Docker build | Image has only one intended migration source/provider; no duplicate manifest artifact |
| Security owner grants | Routine signatures, owner, execute grants/search_path verified |

## Local recovery procedure

1. Mevcut local PostgreSQL volume'larını silme. Önce docker compose ps, operator stderr/stdout, tenant executor logları ve Postgres logları alınır.
2. Operation ID / tenant ID / job ID ile platform.tenancy_onboarding_operations, platform.tenancy_tenant_migration_jobs, attempts, receipts ve tenant goose_db_version state'i read-only incelenir. Secret/table data dump edilmez.
3. Job state succeeded iken onboarding state migrating ise completion→onboarding refresh geçişi ve revision compare incelenir; job tekrar DDL çalıştırmaya zorlanmaz.
4. Job running ise lease, heartbeat ve executor readiness kontrol edilir; aktif/valid lease varken manuel state update yapılmaz.
5. Goose history migration applied gösteriyorsa aynı migration SQL'i düzenlenmez veya yeniden uygulanmaz. State reconciliation için idempotent forward operation/migration tasarlanır.
6. Applied Goose migrations never rewrite; yeni SQL correction yeni timestamp version olarak gelir.
7. Sadece disposable test environment cleanup yapılır. task local:destroy, docker volume prune, DROP DATABASE, DROP SCHEMA ... CASCADE kullanıcı onayı ve açık recovery planı olmadan çalıştırılmaz.
8. Recovery sonrası task dev iki ardışık kez, first tenant + second tenant ve API/worker readiness E2E çalıştırılır.

## Audit rapor şablonu

Her task/phase sonunda şunlar raporlanmalı:

- Branch + base/head commit.
- Değişen path'ler ve her birinin gerekçesi.
- Kaldırılan/taşınan migration framework sembolleri ve consumer listesi.
- Runtime failure stage, operation/job identifiers (sensitive olmayanlar), expected/actual state/revision ve SQLSTATE.
- Goose applied versions (platform ve her tenant ayrı).
- Storage owner + routine grants + SECURITY DEFINER/search_path doğrulama.
- Komutlar, testler, exit code ve CI run linki.
- Local persistent volumes / DB üzerinde destructive action olup olmadığı (varsayılan: hayır).
- Kalan migration ile ilişkili her referansın KEEP/REMOVE/MOVE gerekçesi.
