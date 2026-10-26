# Backend V2 Skill Routing

Komut, kod veya dokuman degisikliginden once bu tabloyu kullan.

| Is | Zorunlu skill'ler |
| --- | --- |
| Her dosya degisikligi | `code-organization` |
| Her source/test/guard degisikligi | `code-organization`, `testing-quality` |
| Yalniz docs/skill/config degisikligi | `code-organization`, ilgili alan skill'i |
| Core runtime, composition, config, telemetry, health | `platform-runtime`, `data-tenancy`, `testing-quality` |
| Deployment mode, installation, Instance Owner, first admin | `platform-runtime`, `data-tenancy`, `iam-system-resources`, `billing`, `events-jobs`, `impact-analysis` |
| V1 adapter, shadow, adoption, fence, rollout | `cutover-compatibility`, ilgili domain skill'i, `impact-analysis` |
| AI, MCP, updater veya yeni deployable | `satellite-services`, `impact-analysis` |
| Yeni modul veya sinir degisikligi | `modular-architecture`, `impact-analysis` |
| Birden fazla modulu etkileyen is | `impact-analysis`, `modular-architecture` |
| Yeni veya degisen endpoint | `api-contract`, ilgili domain skill'i |
| SQL, repository, migration, tenant | `data-tenancy`, `testing-quality` |
| DB diagnostics veya environment recovery set/restore | `platform-runtime`, `data-tenancy`, `documents`, `iam-system-resources`, `events-jobs`, `impact-analysis` |
| Users, roles, organizations, auth, RBAC | `iam-system-resources`, `data-tenancy` |
| Form config, field, publish, record CRUD | `form-engine`, `rest-query`, `data-tenancy` |
| Filter, search, sort, include, aggregate | `rest-query`, ilgili domain skill'i |
| Workflow definition, node veya runtime | `workflow-engine`, `events-jobs` |
| Workflow form verisini etkiliyorsa | `workflow-engine`, `form-engine`, `events-jobs` |
| Outbox, job, worker, RabbitMQ | `events-jobs`, `data-tenancy` |
| Package, install/upgrade/purge, menu | `catalog`, `events-jobs`, `impact-analysis` |
| Board, task, comment, calendar | `work`, `rest-query`, `events-jobs` |
| Document, upload, blob, archive | `documents`, `events-jobs`, `data-tenancy` |
| Billing, entitlement, usage, provider | `billing`, `events-jobs`, `data-tenancy` |
| Saved query, dashboard, publication | `analytics`, `rest-query`, `data-tenancy` |
| Notification, inbox, provider, realtime | `notifications`, `events-jobs` |
| Test, CI veya guard | `testing-quality`, `code-organization` |

## Dosyalar

- `skills/code-organization/SKILL.md`
- `skills/platform-runtime/SKILL.md`
- `skills/cutover-compatibility/SKILL.md`
- `skills/satellite-services/SKILL.md`
- `skills/modular-architecture/SKILL.md`
- `skills/impact-analysis/SKILL.md`
- `skills/api-contract/SKILL.md`
- `skills/data-tenancy/SKILL.md`
- `skills/iam-system-resources/SKILL.md`
- `skills/form-engine/SKILL.md`
- `skills/rest-query/SKILL.md`
- `skills/workflow-engine/SKILL.md`
- `skills/events-jobs/SKILL.md`
- `skills/catalog/SKILL.md`
- `skills/work/SKILL.md`
- `skills/documents/SKILL.md`
- `skills/billing/SKILL.md`
- `skills/analytics/SKILL.md`
- `skills/notifications/SKILL.md`
- `skills/testing-quality/SKILL.md`

Eslesme belirsizse daha az skill okumak yerine ilgili olabileceklerin tamamini oku.
