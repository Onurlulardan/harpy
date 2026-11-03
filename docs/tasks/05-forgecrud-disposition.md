# 05 ForgeCRUD residue disposition

Decisions use the consumer graph. A token match is not a delete or keep decision. `UNRESOLVED` blocks task 04 completion.

## Counts

| Decision | Count |
| --- | ---: |
| REMOVE_FORGECRUD_PRODUCT | 6 |
| KEEP_GENERIC_FOUNDATION | 11 |
| REPLACE_WITH_GENERIC_PRIMITIVE | 0 |
| UNRESOLVED | 616 |

## Rows

| path_or_object | symbol_or_feature | forgecrud_evidence | harpy_consumers | decision | owner | action_or_replacement | test_evidence | reason |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `core/internal/platform/cutover` | cutover adoption, finalization, rehearsal | Task 04 §16 and README §11: cutover_adoption and cutover_finalization are not carried | `core/internal/modules/*/cutover_apply_contract.go`, `core/public/cutover/full_rollout.go` | REMOVE_FORGECRUD_PRODUCT | platform cutover | delete package and contract callers | package tests removed with the package | Cutover is V1 adoption state, not a Harpy generic primitive |
| `**/v1adoption/**` | V1 row adoption SQL | README §2 forbids adoption and compatibility | work, billing, documents, support, usage adapters | REMOVE_FORGECRUD_PRODUCT | each module adapter | delete those trees | adapter unit tests that do not import them | Adoption copies old rows; the fresh Goose schema does not adopt them |
| schema-executor and typed DDL IR | dynamic form DDL | `docs/tasks/05-forgecrud-residue-audit.md` §6 | `tools/localconfig` `schema-executor.env` | REMOVE_FORGECRUD_PRODUCT | localconfig | generator drops the artifact | `profile_environment_test.go` | No Harpy form product; no replacement form store |
| form and workflow designer contracts | `workflow.forms.set-field`, forms module imports | Task 05 §3.2 search list | `core/internal/platform/transaction` tests, archguard fixture | REMOVE_FORGECRUD_PRODUCT | transaction tests and archguard fixture | remove product operation names from fixtures | transaction unit tests | Designer behavior is not a generic Harpy requirement |
| `core/internal/modules/tenancy/application/tenantmigration` | tenant migration job enqueue | Task 03 §8 one-shot Goose container | `environmentoperator.migrate`, `tenancy.OperatorRuntime` | REMOVE_FORGECRUD_PRODUCT | tenancy | `provisionlaunch.Run` then onboarding accept | `environmentoperator` initialize tests | The job is an executor checkpoint, not tenant identity |
| `core/internal/platform/schemacontract` | binary migration manifest gate | Task 04 §2: startup does not read a schema contract | postgres integration fixtures, `tools/localrelease` | REMOVE_FORGECRUD_PRODUCT | platform postgres | connection and role readiness only | `open_pool_api_test.go` | Manifest digest is migration state |
| IAM identity and authorization | users, roles, sessions | `docs/architecture/06` and `skills/iam-system-resources` | API and worker composition | KEEP_GENERIC_FOUNDATION | iam | keep module; table DDL already in Goose | iam unit tests | Generic identity, not a product preset |
| Tenancy registry and onboarding | tenant id, schema name, installation delivery | `docs/architecture/06` and `skills/data-tenancy` | operator initialize, customer provisioning | KEEP_GENERIC_FOUNDATION | tenancy | keep onboarding and license delivery | operator initialize tests | Installation and tenant identity stay |
| Billing license and subscription | offering, entitlement, invoice | `docs/architecture/15` §3 and `skills/billing` | billing HTTP runtime | KEEP_GENERIC_FOUNDATION | billing | keep module | billing unit tests | Platform license is a Harpy requirement |
| Work boards and tasks | board, column, task | `docs/architecture/06` and `skills/work` | work HTTP | KEEP_GENERIC_FOUNDATION | work | keep module; `work_*` Goose tables | work domain tests | Current work contract, not a designer |
| Documents | folder, item, upload | `docs/architecture/06` and `skills/documents` | documents HTTP and jobs | KEEP_GENERIC_FOUNDATION | documents | keep module; `documents_*` Goose tables | documents unit tests | Generic object storage |
| Notifications | message, delivery, preference | `docs/architecture/06` and `skills/notifications` | notifications HTTP | KEEP_GENERIC_FOUNDATION | notifications | keep module | notifications unit tests | Generic delivery |
| Analytics saved query and dashboard | query version, dashboard version | `docs/architecture/06` and `skills/analytics` | analytics HTTP | KEEP_GENERIC_FOUNDATION | analytics | keep module | analytics unit tests | Listed Harpy domain, not a new designer |
| Audit entries | append-only entry | `docs/architecture/06` §8 | audit HTTP | KEEP_GENERIC_FOUNDATION | audit | keep module | audit unit tests | Generic audit |
| Platformops recovery | backup, restore, signing | Task 04 §16 | `core/cmd/recovery` | KEEP_GENERIC_FOUNDATION | platformops | `core-recovery` | `composition/recovery` tests | Recovery is not migration |
| HTTP, config, telemetry, health | runtime primitives | Task 05 §4 | API and worker | KEEP_GENERIC_FOUNDATION | platform | keep | composition tests | Generic foundation |
| Historical `*_v1` routines called by KEEP modules | stored procedure bodies | Task 03 §14.1 forbids copying the signature | `docs/tasks/05-v1-routine-disposition.md` lists 616 SQL files | UNRESOLVED | each KEEP module | do not create the function; table DML only where the Go result rule is already typed | each path in that file | The Go call names the routine, and the result rule is not determined without the deleted dump body |
