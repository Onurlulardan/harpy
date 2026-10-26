---
name: work
description: Build backend-v2 boards, columns, memberships, tasks, labels, checklists, comments, calendar events, recurrence, unified inbox/calendar projections, and Work V1 adoption. Use for task-manager, board, comment, assignment, ordering, reminder, recurrence, or calendar behavior.
---

# Work

Normatif domain karari icin ADR-0016 ve ADR-0029'u; purge/reference varsa
ADR-0026'yi eksiksiz oku. ADR ile bu skill celisirse ADR authoritative'dir; yeni semantik
gerekiyorsa source yazmadan Accepted ADR olustur.

## Baslangic kapisi

1. Degisen aggregate, owner state, actor ve expected row version'i yaz.
2. `modular-architecture`, `impact-analysis` ve `code-organization` skill'lerini oku.
3. HTTP icin `api-contract`; SQL icin `data-tenancy`; liste/range icin
   `rest-query`; principal icin `iam-system-resources`; reminder/recurrence ve
   projection icin `events-jobs` oku.
4. Forms veya Workflow link/projection etkileniyorsa ilgili domain skill'ini oku.
5. V1 adoption/cutover, permission revoke ve rollback senaryosunu yaz.
6. Gecici task modeli, JSONB fallback veya ikinci canonical state varsa dur.

## Ownership siniri

Work su business state'in tek owner'idir:

- board, column, USER/ROLE/ORGANIZATION membership ve board role;
- Work task, assignee, label, checklist ve comment;
- manual calendar event ve recurrence rule;
- Work resource link'i.

Workflow approval/human task'i Workflow owner'inda kalir. Onu Work task olarak
kopyalama; unified inbox/calendar'da source ID/version tasiyan rebuild edilebilir
projection kullan. Form, Workflow veya IAM tablosuna SQL, FK, JOIN, cascade ya da
direct delete yazma. Existence, summary ve command icin owner public API'sini
kullan. Cross-module link'te slug/title/table name degil stable typed ID kullan.

Projection'i source of truth veya permission grant'i yapma. Detail/action'i
owner'a delege et ve her read'de current permission revision'i kontrol et. Stale
veya yetkisiz item'i gizle; persisted repair job'i ac.

## Aggregate kurallari

- Effective board role'u current IAM principal setinden hesapla; global IAM
  action ile aggregate role'un ikisini de application command'da zorunlu tut.
- Son aktif USER/ADMIN'i dusurme veya silme.
- Assignee'nin ayni board'da en az VIEWER access'i yoksa assignment'i reddet.
- Move komutunda source/target column'u deterministik sirada lock et, neighbor
  ownership'ini dogrula, status ve contiguous position'i tek commit'te yaz.
- Mutation'da expected row version ve DB-backed idempotency receipt'i uygula.
- DONE/CANCELLED terminal gecislerini explicit complete/cancel/reopen command'lari
  disinda degistirme.
- Label/checklist'i typed relational tut; arbitrary JSONB business modeli acma.
- Comment update yapma; author redaction ve moderator redaction'i tombstone/audit
  semantigiyle ayir.
- Archive/soft delete/restore ile purge'u ayir; hard delete'i generic CRUD'a koyma.

Calendar'da timed ve all-day modellerini karistirma. RFC 3339 UTC, PostgreSQL
date, IANA TZID ve ADR-0012 TZDB/DST pinning sozlesmesini uygula. Recurrence'i
versioned RFC 5545 parser ile validate et; occurrence'i DB job ile materialize
et ve `(rule_id, occurrence_start)` uniqueness ile duplicate'i engelle.

## Dosyalama

Resource ve operation bazinda ayir:

```text
modules/work/application/
  boards/create.go       boards/create_test.go
  boards/archive.go      boards/archive_test.go
  memberships/add.go     memberships/add_test.go
  tasks/create.go        tasks/create_test.go
  tasks/move.go          tasks/move_test.go
  tasks/complete.go      tasks/complete_test.go
  comments/create.go     comments/create_test.go
  comments/redact.go     comments/redact_test.go
  calendar/create.go     calendar/create_test.go
  recurrence/materialize.go
```

Move, recurrence materialization, assignment ve comment lifecycle'i ayri
use-case'tir. Dosya/CRUD/isim siniri icin
`skills/code-organization/SKILL.md` tek authoritative kaynaktir. Gecici rank,
storage, transport veya compatibility domain modeli kurma.

## API, tenant, auth ve REST Query

- Tenant'i verified actor context/TenantScope'tan al; client/header tenant ID'sini
  row alanina yazma.
- IAM action ve effective board/resource visibility'yi hem query hem command'da
  uygula; list/detail farkinda fail-open dal acma.
- Tum mutation'larda `Idempotency-Key`; aggregate mutation'inda expected version
  iste. State, audit, outbox ve job intent'ini ayni transaction'da yaz.
- List, filter, search, sort, inbox ve calendar range icin yalniz REST Query
  `QuerySpec` kullan. Include'i Work-owned declared relation'larla sinirla.
- IAM/Forms/Workflow enrichment'i batch public API veya projection ile yap;
  request basina cross-module N+1 ve owner-disi JOIN yazma.
- Network, notification veya broker publish'i transaction icinde calistirma.

## Jobs ve cutover

Reminder, recurrence, projection repair ve purge ADR-0011 DB job'laridir;
RabbitMQ yalniz wake-up/transport'tur. Consumer inbox receipt, revision gap,
duplicate, reorder ve rebuild checkpoint'ini uygula. Notification'i commit
sonrasi idempotent external effect olarak tut.

ADR-0013 single-writer protokolunu `(tenant,work)` biriminde uygula. V1 state'i
inactive Work tablolarina stable ID/checksum ile backfill et; invalid enum,
cross-board reference, orphan, unresolved slug veya recurrence'i
`WORK_ADOPTION_BLOCKED` yap. `workflow_tasks` Workflow'da kalir. Production
dual-write veya iki canonical task modeli acma.

Compatibility adapter'ini edge'de, persisted cohort/removal gate'iyle tut. V2
write sonrasi V1 writer donusunu normal rollback sayma; reverse reconciliation
ve sifir fark olmadan acma.

## Verification

- Board role/principal matrix, last-admin ve IAM revoke concurrency testleri.
- Cross-board/cross-tenant ID substitution ve owner-disi SQL/import testleri.
- Move lock/order, stale version, retry ve status-transition testleri.
- Comment redaction, archive/restore ve purge blocker/cooling/legal-hold testleri.
- RFC 5545, TZDB/DST, duplicate occurrence, lease expiry ve worker crash testleri.
- Projection duplicate/gap/revoke/rebuild ve delegated-action authorization.
- REST Query field permission, budget, stable pagination ve range boundaries.
- V1 differential/backfill/checksum, writer fence ve rollback tatbikati.
- Operation-eslesmeli testler, architecture testleri ve fileguard.
