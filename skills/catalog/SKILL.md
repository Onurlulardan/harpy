---
name: catalog
description: Govern backend-v2 package manifests, module discovery, dependency planning, tenant installation lifecycle, menu versioning, demo receipts, and Catalog V1 adoption. Use for install, upgrade, enable, disable, restore, retire, purge, menu publish, package registry, or module-catalog compatibility work.
---

# Catalog

Normatif domain karari icin once
ADR-0015, ADR-0026 ve ADR-0032'yi eksiksiz oku. ADR ile bu
skill celisirse ADR authoritative'dir; yeni semantik gerekiyorsa source yazmadan
Accepted ADR olustur.
Instance/tenant authority veya frontend visibility etkileniyorsa ADR-0117 ve
`docs/phase14-platform-capability-matrix.md` dosyasini da oku.
Standalone migration plan/apply/run UI/API acma; migration evidence'i
`catalog.upgrade` operation plan/receipt'i ve ilgili owner artifact'inda kalir.

## Baslangic kapisi

1. Degisiklik owner'ini, tenant/package lifecycle state'ini ve public command'i yaz.
2. `modular-architecture`, `impact-analysis` ve `code-organization` skill'lerini oku.
3. HTTP icin `api-contract`; SQL icin `data-tenancy`; liste icin `rest-query`;
   IAM grant'i icin `iam-system-resources`; async is icin `events-jobs` oku.
4. V1 adoption veya route degisiyorsa `(tenant,catalog)` cutover, client
   capability, rollback ve otomatik compatibility-exit gate'ini planla.
5. Acik domain karari, gecici state modeli veya ikinci yazma yolu varsa dur.

## Ownership siniri

Catalog yalniz sunlarin owner'idir:

- signed immutable package/version registry;
- dependency resolution ve tenant installation state'i;
- component binding ve durable lifecycle operation state'i;
- demo-data receipt'i;
- canonical menu definition, version, item ve tenant override'i.

Forms, Workflow, IAM, Analytics ve diger component owner'larinin tablosuna SQL,
DDL, FK, cascade veya adapter importu yazma. Component'i owner public API'siyle
stage, verify, activate, disable, restore veya purge et. Slug'i identity/storage
key'i yapma; stable typed ID ve exact version kullan.

Manifest'i guvenilmeyen declarative program olarak validate et. Unknown kind,
unsupported contract, cycle, ID collision, hash veya signature hatasini
fail-closed reddet. Runtime package upload'i ya da unsigned manifest yolu acma.

## Lifecycle uygulama akisi

- Install/upgrade planinda dependency range'lerini exact version'a coz ve plan
  hash'ini operation state'ine dondur.
- Component'leri inactive stage et; uzun DDL/backfill'i nihai schema executor,
  diger uzun adimlari DB job uzerinden calistir.
- Tum owner readiness/hash kontrollerinden sonra pointer, IAM grant, Catalog
  state/menu, audit ve outbox'i ayni TenantScope commit'inde aktive et.
- Retry'yi persisted operation step/receipt'ten devam ettir; tahmin veya
  best-effort cleanup yapma.
- Disable veriyi silmez; retire inactive tutar; restore explicit command'dir.
- Purge ADR-0015 approval, step-up, cooling, blocker, receipt ve tombstone
  kontratlarinin tamami green olmadan baslamaz.
- Downgrade, direct seed, sync-install veya synchronous destructive uninstall
  adinda ikinci lifecycle yolu acma.

Menu item'ini raw route/slug ile degil typed target ile bagla. Tenant override'ini
package defaultundan ayir ve upgrade'de koru. Menu visibility'yi authorization
sayma; response'u IAM ve target-owner permission'iyle yeniden filtrele.

## Dosyalama

Her operation'i ayri source ve test dosyasinda tut:

```text
modules/catalog/application/
  install.go        install_test.go
  upgrade.go        upgrade_test.go
  enable.go         enable_test.go
  disable.go        disable_test.go
  restore.go        restore_test.go
  retire.go         retire_test.go
  purge.go          purge_test.go
  publish_menu.go   publish_menu_test.go
```

Install, upgrade, enable, disable, rollback, retire ve purge birbirinden ayri
use-case'tir. Dosya/CRUD/isim siniri icin `skills/code-organization/SKILL.md`
tek authoritative kaynaktir. Gecici package veya lifecycle modeli kurma.

## API, tenant, auth ve query

- Actor ve tenant'i verified context/TenantScope'tan al; header veya manifest'i
  tenant kaynagi yapma.
- Her mutation'da ADR-0015'in ayri IAM action'ini application command icinde
  reauthorize et; manifest permission template'ini grant kabul etme.
- Lifecycle mutation'larinda `Idempotency-Key`, persisted receipt ve stable
  operation handle kullan.
- List/filter/search/sort icin yalniz typed REST Query `QuerySpec` kullan;
  ad-hoc filter map'i, raw identifier, cross-owner include/JOIN yazma.
- Operation, audit, outbox ve gerekli job'u transaction kontratina gore atomik
  kaydet; RabbitMQ'yu state veya scheduler yapma.

## Cutover

ADR-0013 single-writer protokolunu `(tenant,catalog)` biriminde uygula.
Production dual-write/sync acma. V1 adoption'i stable ID ve checksum ile inactive
state'e yap; unresolved target/orphan/hash farkini `CATALOG_ADOPTION_BLOCKED`
olarak persisted ve operator-visible tut. Mutation cutover'unu gerekli owner
domainleri V2 olduktan ve signed client capability green olduktan sonra ac.

Compatibility adapter'ini edge'de tut. Persisted cohort/gate owner'i, sifir
legacy-hit kosulu ve architecture-test removal gate'i olmadan gecici route
ekleme. V2 write sonrasi V1 writer donusunu normal rollback gibi sunma.

## Verification

- Manifest signature/hash, unknown component, collision ve dependency-cycle.
- Install/upgrade duplicate, crash, stale plan, resume ve failed activation.
- Required/optional dependency ile disable/retire/restore/purge blocker matrix'i.
- Menu override, permission filtering, CAS race, cache rebuild ve target revoke.
- Tenant substitution, IAM escalation ve owner-disi SQL/import negative testleri.
- Outbox/job duplicate, lease expiry, DLQ ve owner receipt idempotency testleri.
- V1 backfill/checksum, client capability, writer fence ve rollback tatbikati.
- Operation-eslesmeli testler, architecture testleri ve fileguard.
