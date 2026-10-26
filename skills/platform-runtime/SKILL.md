---
name: platform-runtime
description: Build backend-v2 composition, HTTP runtime, configuration, telemetry, tenant transaction, database roles, migration runtime, and health surfaces. Use for core-api, core-worker, core-migrator, platform packages, bootstrap, or deployment wiring.
---

# Platform Runtime

## Zorunlu okuma

1. `docs/architecture.md`
2. `docs/adr/0001-modular-core.md`
3. `docs/adr/0004-tenant-transaction-scope.md`
4. `docs/adr/0005-platform-runtime.md`
5. `docs/adr/0006-tenancy-migrations.md`
6. `docs/adr/0007-architecture-enforcement.md`
7. `docs/adr/0033-platform-runtime-wire-and-config.md`
8. `docs/adr/0034-postgres-runtime-tenantscope-schema-contract.md`
9. `docs/adr/0116-environment-deployment-profile-and-tenant-capacity.md`
10. `docs/adr/0118-installation-onboarding-operator.md`
11. SaaS customer provisioning etkileniyorsa
    `docs/adr/0119-saas-customer-tenant-provisioning.md`
12. License/recovery/protected authority etkileniyorsa ADR-0120, ADR-0121 ve ADR-0123..0127
13. Phase 14 process/config wiring'inde `docs/phase14-runtime-profile-config.md`
14. Bootstrap authority/opening usage etkileniyorsa ADR-0122

## Deployable siniri

- `core-api`, `core-worker` ve `core-migrator` ayni modular core'u compose eder.
- `core-operator` yalniz `initialize`, `reconcile-restored-identity` ve
  `reconcile-restored-license` exact one-shot ABI'laridir.
- New-tenant DDL yalniz `core-migrator tenant-migration-executor` profilindedir.
- Recovery-set DB backup/restore yalniz ADR-0121/0123 isolated migrator profillerindedir.
- Forms/Workflow servis degil module'dur; async processor'lari worker'da calisir.
- Composition yalniz constructor/wiring yapar; business policy tasimaz.
- Core ici modul iletisiminde HTTP veya broker kullanma.
- Runtime process'ine DDL, extension-admin veya adopter credential'i verme.

## Request sirasi

1. Stable request/correlation ID olustur veya dogrula.
2. Authenticator ile actor'u cozumle.
3. Tenant membership ve route authority'sini dogrula.
4. Application command'i TenantScope transaction'ini acar.
5. Module command/query portunu cagirir.
6. Commit sonrasi canonical envelope ve telemetry tamamlanir.

Middleware transaction baslatmaz, tenant header'ini authority saymaz ve handler
business transition secmez.

## Config ve gozlemlenebilirlik

- ADR-0005 typed config schema'si disinda environment lookup yapma.
- Secret degerini log, metric label, trace attribute veya error body'ye koyma.
- Request, tenant, actor, operation, module ve cutover generation context'ini
  redaction policy'sine gore propagate et.
- Health endpoint'i dependency durumunu saklamaz; readiness fail-closed olur.
- Metric label cardinality'sine tenant/user/form ID koyma.

## Database ve migration

- PostgreSQL runtime yalniz ADR-0034 direct endpoint ve exact image digest'iyle
  kurulur; PgBouncer veya gecici pool adapter'i ekleme.
- Her process profile'inda exact `session_user/current_user` preflight'i yap;
  privileged bootstrap/extension credential'ini normal runtime'a tasima.
- Checkout/checkin state'ini denetle; tainted connection'i destroy et ve
  `DISCARD ALL` ile yeniden kullanima alma.
- Tenant SQL yalniz TenantScope icinde ve fully-qualified platform erisimiyle.
- Tenant'i verified actor'dan turet; caller/header tenant argumani kabul etme.
- Tenant registry, schema sentinel ve authority group/domain lock'larini callback
  oncesi al; success'i yalniz commit tamamlandiktan sonra uret.
- Runtime login DDL yapamaz; migrator typed manifest/IR disina cikamaz.
- Normal migrator privileged islemi calistirmaz; signed bootstrap/extension
  receipt'lerini dogrulamadan foundation DDL'ine baslamaz.
- Runtime/database generation ve manifest hash'leri schema-contract readiness
  denklemini gecmeden trafik, claim veya schema lease acma.
- Startup migration yoktur; deploy migration'i ayri one-shot authority'dir.

## Tamamlama kapisi

- Composition import graph'i archguard'dan gecti.
- Config missing/invalid/secret-redaction testleri var.
- Health, graceful shutdown ve in-flight request testleri var.
- Iki tenant negative isolation ve pool reuse testleri var.
- Exact-digest gercek PG18.4 integration gate'i Docker yokken skip edilmiyor.
- Pool taint, sentinel, stale authority generation, signed receipt ve rolling
  schema-contract negative testleri var.
- Runtime/migrator/adopter privilege negative testleri var.
- Ilgili ADR verification maddeleri executable test veya runbook'a bagli.
