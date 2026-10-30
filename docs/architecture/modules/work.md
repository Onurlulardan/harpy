# Modül — Work

## 1. Owner alanı

Work, iş panosu, görev, takvim ve bu nesnelerin collaboration ilişkilerinin owner'ıdır. Temel entity'ler board, column, membership, task, checklist, label, comment, calendar event ve recurrence rule'dur.

Kaynak kök: core/internal/modules/work/

## 2. Board ve column

Board:
- ID ve CreatorID
- Name/Description
- Lifecycle: ACTIVE veya ARCHIVED
- RowVersion ve CreatedAt

Column:
- BoardID
- Name/Color
- TaskStatus
- Sparse position
- RowVersion

CreateBoard use-case'i yeni board, initial column ve creator için ADMIN membership oluşturur. Bunların tamamı aynı Work mutation boundary içinde persist edilir.

## 3. Task domain

Task kind:
- CUSTOM
- FORM_FILL
- FORM_REVIEW

Status:
- OPEN
- IN_PROGRESS
- IN_REVIEW
- DONE
- CANCELLED

Priority:
- LOW
- MEDIUM
- HIGH
- URGENT

Domain NextStatus transition fonksiyonu status geçişlerini doğrular:
- MOVE terminal task üzerinde yapılamaz ve target CANCELLED olamaz.
- COMPLETE yalnız DONE'a geçebilir.
- CANCEL yalnız CANCELLED'a geçebilir.
- REOPEN yalnız terminal state'ten OPEN'a döner.
- DONE ve CANCELLED terminal statülerdir.

Task transition business invariant'ı application/HTTP katmanına kopyalanmaz; domain fonksiyonu tarafından uygulanır.

## 4. Task application

Tasks application create, update, move, complete, cancel, reopen, assign, label, checklist ve resource link gibi use-case'leri ayrıştırır.

CreateTask:
1. Input normalization ve type validation yapar.
2. Resource kind ile resource link uyumunu kontrol eder.
3. Assignee principal ve permission revision'larını doğrular.
4. Request fingerprint üretir.
5. Mutation engine üzerinden idempotency ve transaction sınırına girer.
6. Board ve column'u for-update yükler.
7. Board active ve column'un aynı board'a ait olduğunu doğrular.
8. Task, relation ve event/audit state'ini persist eder.

MoveTask expected RowVersion ister. Task ve ilgili column'lar kilitlenir. Position, mutlak index yerine before/after neighbor ID üzerinden hesaplanır. Eski order input'u yalnız compatibility decode için normalize edilir.

## 5. Membership ve authorization

Board membership USER/ROLE/ORGANIZATION principal kavramını kullanır. Board role ve active membership ile birlikte değerlendirilir. Work authorization IAM principal resolver ve permission verifier ile composition'da bağlanır.

Task assignee, label, checklist, resource link ve board membership değişiklikleri her biri ayrı use-case ve validation yüzeyine sahiptir. User'ın bir board'u görebilmesi, tüm task resource link'lerini okuyabileceği anlamına gelmez.

## 6. Calendar ve recurrence

Calendar event audience:
- PRIVATE
- BOARD
- ORGANIZATION

CalendarWindow ya timed start/end UTC instant, ya da all-day start/end date taşır. Aynı anda iki format birden kullanılamaz. Timezone zorunludur.

Event recurrence generated ise RecurrenceRuleID ve RecurrenceRevision ikisi birlikte geçerli olmalıdır. Manuel event recurrence rule'a bağlı değildir.

Recurrence contract work_rrule_v1 biçimindedir. Desteklenen frequency DAILY, WEEKLY, MONTHLY, YEARLY'dir. Parser:
- Timezone ve pinned TZDBVersion ister.
- RRULE token'larını strict parse eder.
- Duplicate token ve invalid value'ları reddeder.
- Count/interval upper bound uygular.
- EXDATE sayısını sınırlar.
- DST ambiguous/nonexistent local time değerlerinde deterministic resolution ister.

MaterializeRecurrence service rule revision'ı doğrular, doğru tzdb version'ını yükler, occurrence aralığını hesaplar ve future occurrence limit'lerine uyar.

## 7. Background jobs

Work worker jobs şu tip işleri kapsar:
- Reminder notification intent
- Recurrence occurrence materialization
- Scheduled calendar processing
- Projection/cleanup görevleri

Worker job yalnız Work application/API sözleşmesini çağırır; Work domain logic worker composition içine kopyalanmaz.

## 8. Modüller arası resource links

Task resource link bir owner module ve target listesi taşır. Target type, target ID ve version ID doğrulanır. Permission revision/fingerprint idempotency digest'e katılır. Böylece task'a bağlı bir dış kaynağın izni veya sürümü değiştiğinde eski authorization sonucuna güvenilmez.

## 9. Kaynak haritası

- Domain: core/internal/modules/work/domain/
- Application: core/internal/modules/work/application/boards/, tasks/, calendar/, recurrence/, memberships/, comments/, labels/, checklists/
- Ports: core/internal/modules/work/ports/
- HTTP: core/internal/modules/work/adapters/http/ ve module HTTP handler dosyaları
- PostgreSQL: core/internal/modules/work/adapters/postgres/
- Composition: core/internal/composition/api/build_work_http.go, worker/build_work_jobs.go
- Contract: contracts/authority/work_*.json, contracts/events/work-*.schema.json, contracts/http/work*
- Skill: skills/work/SKILL.md
