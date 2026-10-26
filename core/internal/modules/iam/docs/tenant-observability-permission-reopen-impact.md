# IAM tenant observability permission reopen impact

## Scope

This controlled reopen is limited to `core/internal/modules/iam/**`. It adds a
paired tenant migration and IAM-local tests for the permission rows already
required by the frozen Audit and Usage IAM verifier contracts. It does not edit
composition, shared manifests, operation registries, OpenAPI, other modules or
legacy source.

## Proven runtime gap

- `audit_access_permission.go` accepts five operations only when the permission
  is exactly `changelog.read`, mapped to resource `changelog`, action `read`.
- `usage_access_permission.go` accepts eight tenant operations only when the
  permission is exactly `analytics.usage.read`, mapped to resource
  `analytics.usage`, action `read`.
- Neither exact tenant permission is published by the current IAM tenant
  migration stream. The verifier therefore fails closed at a real database
  authorization check.

## Legacy grant evidence

The V1 gateway protects all five changelog routes with
`RequirePermission("changelog", "read")` in `backend/api-gateway/main.go`.
The V1 default permission matrix in `backend/shared/database/seeder.go` grants
that pair to exactly these default roles:

- `Tenant Yöneticisi`
- `Genel Müdür`
- `Temel Kullanıcı`

The V1 tenant usage routes accept `traces.read` or `traces.manage`. In the same
default matrix only `Tenant Yöneticisi` receives `traces.read`; neither `Genel
Müdür` nor `Temel Kullanıcı` receives trace authority. Therefore
`analytics.usage.read` is granted only to the tenant administrator mapping.

V1 cleanup evidence defines only these legacy aliases:

- `Yönetici` -> `Tenant Yöneticisi`
- `Müdür` -> `Genel Müdür`
- `Kullanıcı` -> `Temel Kullanıcı`

The migration may recognize those evidenced aliases. It must not infer English
aliases, grant all authenticated roles, copy platform super-admin authority, or
turn either permission into a wildcard.

## Data and authorization impact

- Owner: IAM; scope: tenant schema; rollout: paired `fresh` and `v1_adopted`.
- Add resources `changelog` (owner `audit`) and `analytics.usage` (owner
  `usage`), reuse/publish action `read`, and publish two exact permissions.
- Add or reactivate exact role-permission bindings only for the evidenced role
  mapping above. The protected tenant administrator remains covered by its
  existing `tenant_all` policy; no redundant explicit binding is created.
- Bump the tenant authorization clock and active user authorization versions so
  stale sessions cannot retain pre-migration authorization evidence.
- No user or organization binding is synthesized. Custom-role compatibility is
  not guessed in this slice and remains subject to explicit adoption evidence.

## Verification boundary

The focused contract test must prove byte-identical origins, source-policy
compliance, exact resources/actions/permissions and exact role mapping. The
PostgreSQL 18 test must apply both origins to fresh IAM foundations, exercise
canonical plus evidenced alias roles, prove non-evidenced roles receive no
grant, and prove protected tenant-all authorization still works.
