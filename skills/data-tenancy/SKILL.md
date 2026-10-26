---
name: data-tenancy
description: Govern backend-v2 PostgreSQL ownership, schema-per-tenant isolation, Unit of Work, repositories, migrations, backfills, indexes, transactions, and data rollout. Use for every SQL, model, repository, tenant lifecycle, or migration change.
---

# Data and Tenancy

Kesin PostgreSQL, schema, role ve migration kontrati icin
`docs/adr/0006-tenancy-migrations.md` dosyasini da oku.
Deployment capacity/environment onboarding icin ADR-0116 ve ADR-0118'i; SaaS
customer provisioning icin ADR-0119'u da oku.
DB diagnostics icin ADR-0121'i; recovery/restore/head icin ADR-0121 ve ADR-0123..0127'yi,
process key ownership'i icin `docs/phase14-runtime-profile-config.md`yi oku.

## Veri sinirini belirle

Her degisiklikte sunlari yaz:

- Platform schema mi tenant schema mi?
- Owner modul ve owner migration hangisi?
- Command transaction'i nerede baslatir?
- Mevcut tenant'lar nasil upgrade edilir?
- Rollback image/route ile mi, forward-fix ile mi yapilir?

## Schema-per-tenant

- Tenant schema adini client slug'indan request basina uretme.
- Server-side kaydedilen immutable `schema_name` kullan.
- Tenant transaction search path'ini yalniz `pg_catalog`, tenant schema sirasinda
  kur; runtime TEMP privilege'i veya `pg_temp` kullanmasin.
- Platform tablolarini fully-qualified kullan.
- Bir transaction'i iki tenant schema'sina acma.
- Unit of Work basinda schema identity ile tenant ID'yi dogrula.

## Unit of Work

```text
application command
  -> Tenant Unit of Work
      -> repositories
      -> business state
      -> audit
      -> outbox/job/receipt
  -> commit
  -> HTTP success
```

- Transaction'i HTTP middleware acmasin.
- Repository global DB veya tenant slug almasin; yalniz `DBTX` alsin.
- Commit basarili olmadan success response yazma.
- Network, RabbitMQ, SMTP, AI veya blob upload'i transaction icinde yapma.
- Race riski icin row lock, unique constraint veya optimistic version kullan.

## Ownership

- Bir tabloyu tek modul ve tek migration stream sahiplenir.
- Baska modul tablosuna SQL veya JOIN yazma.
- Modul ici FK kullan; cross-module cascade kullanma.
- Read ihtiyacinda owner query API'si veya rebuild edilebilir projection kullan.
- Redis/cache'i source of truth yapma.

## Migration

- Migration'i yalniz `core-migrator` calistirsin.
- API ve worker startup migration calistirmasin.
- Applied migration'i degistirme; yeni forward migration yaz.
- Her migration'a checksum ve owner module bilgisi ata.
- DDL ve modul kurulumu icin advisory lock kullan.
- Tenant basina checkpoint ve acik failure state tut.
- Existing schema'yi otomatik drop etme.
- Uninstall disable eder; purge ayri, onayli ve audit'li use-case'tir.

## Rollout deseni

```text
expand -> resumable backfill -> application switch -> soak -> contract
```

- Legacy ve V2 ayni anda calisirken additive/nullable schema kullan.
- Production rollback'i destructive down migration'a baglama.
- Backfill'i tenant bazli, idempotent, batch ve checkpoint'li yap.
- Schema contract'i rollback penceresi bittikten sonra uygula.

## Dynamic form DDL

- Public slug'i fiziksel tablo veya kolon kimligi yapma.
- Stable form ve field ID mapping kullan.
- Config publish'te validate ve compile edilmis DDL plan uret.
- Restricted typed DDL IR'yi yalniz `core-migrator schema-executor` ile uygula;
  onceki active version'i failure'da koru.
- Normal form record'u ADR-0010 typed physical storage'inda tut; EAV veya tek
  JSONB generic fallback acma.

## Index ve query

- Index'i gercek REST Query access pattern'iyle gerekcelendir.
- Her `contains/search` operatorunun type/index maliyetini yaz.
- Interactive query'ye statement timeout uygula.
- Exact count ve deep offset'i varsayilan yapma.
- Query plan regression testini kritik form fixture'lariyla calistir.

## Verification

- Fresh database migration.
- En az bir eski snapshot upgrade.
- Iki tenant negatif isolation testi.
- Concurrent mutation/race testi.
- Backfill resume ve retry testi.
- Rollback/forward-fix tatbikati.
- Owner disi tablo erisimi architecture testi.
