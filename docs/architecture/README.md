# Harpy architecture documentation

## Purpose

This documentation describes the architecture implemented in the source tree for developers, technical leads, and operations engineers. It is the canonical technical guide to Harpy; the repository overview and getting-started instructions live in the [root README](../../README.md).

## Reading order

| Order | Document | Scope |
| --- | --- | --- |
| 1 | [General overview](01-genel-bakis.md) | Product scope, technology, repository map |
| 2 | [Architecture and module boundaries](02-mimari-ve-modul-sinirlari.md) | Modular monolith, layers, dependency rules |
| 3 | [Runtime and composition](03-runtime-ve-composition.md) | Setup, API, worker, operator, process lifecycle |
| 4 | [Platform foundations](04-platform-temelleri.md) | Config, PostgreSQL, transactions, telemetry, health |
| 5 | [HTTP, API, and contracts](05-http-api-ve-kontratlar.md) | Route assembly, API contracts, error model, REST Query |
| 6 | [Data and tenant isolation](06-veri-ve-tenant-izolasyonu.md) | Schema-per-tenant, scope, data ownership, Unit of Work |
| 7 | [Events, jobs, and workers](07-event-job-ve-worker.md) | Outbox, RabbitMQ, DB jobs, retries, leases |
| 8 | [Security model](08-guvenlik-modeli.md) | Actor, IAM, RBAC, tokens, tenant security, secrets |
| 9 | [Deployment and operations](09-deployment-ve-operasyon.md) | Compose, profiles, local stack, health, logs, backup |
| 10 | [Testing and engineering standards](10-test-ve-muhendislik-standartlari.md) | Test types, CI, fileguard, archguard |
| 11 | [Developer guide](11-gelistirici-rehberi.md) | Setup, commands, contribution workflow |
| 12 | [Shared capabilities](12-paylasilan-capabilities.md) | REST Query, artifacts, tenant activation, orchestration |
| 13 | [Contract catalog](13-kontrat-katalogu.md) | HTTP, event, runtime, authority, storage contracts |
| 14 | [Architecture decisions](14-mimari-kararlar.md) | Modular monolith and key trade-offs |
| 15 | [Module package map](15-modul-paket-haritasi.md) | Application and adapter package groups |

## Module documentation

| Module | Scope |
| --- | --- |
| [IAM](modules/iam.md) | Identity, authentication, sessions, RBAC, organizations |
| [Tenancy](modules/tenancy.md) | Tenant identity, provisioning, onboarding, settings |
| [Billing](modules/billing.md) | Money, offering, subscription, invoice, entitlement, usage |
| [Catalog](modules/catalog.md) | Signed packages, dependency resolution, component lifecycle, menu |
| [Work](modules/work.md) | Board, task, membership, calendar, recurrence |
| [Documents](modules/documents.md) | Folder, item, version, upload, encryption, scan, archive |
| [Notifications](modules/notifications.md) | Inbox, preference, template, provider, delivery |
| [Analytics](modules/analytics.md) | Query, dashboard, publication, ACL, execution |
| [Other modules](modules/diger-moduller.md) | Audit, Usage, Support, System Resource Layout, Platform Authority/Ops |

## Primary engineering references

- `AGENTS.md`: repository workflow and change boundaries.
- `skills/INDEX.md`: required skill routing by work type.
- `skills/modular-architecture/SKILL.md`: module ownership and dependency direction.
- `skills/code-organization/SKILL.md`: file organization and size standards.
- `skills/testing-quality/SKILL.md`: test selection and quality gates.
- `contracts/`: HTTP, event, runtime, and authority contracts.
- `core/internal/composition/`: process wiring and runtime assembly.

## Documentation accuracy

Module descriptions are based on actual source directories, domain types, application use cases, adapter groups, and test presence. Normative rules in skill files are distinguished from behavior directly observed in source code. When a module's complete use-case inventory is not listed here, use its source tree as the definitive reference.
