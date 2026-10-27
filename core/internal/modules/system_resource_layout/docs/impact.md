# System-resource layout freeze impact

## Ownership

- Owner: `system_resource_layout`.
- Owned data: tenant-scoped layout revisions, publications, field-reference
  evidence and installation-baseline evidence.
- Cross-module inputs: IAM public system-resource field catalog and Forms public
  presentation types only. IAM values, permissions and profile schema remain
  IAM-owned.
- Transaction boundary: the caller opens one verified `TenantScope`; the layout
  runtime binds its own store and never opens a second transaction.

## Contracts and consumers

- Canonical surface: seven revision/publication/presentation HTTP operations.
- Compatibility surface: the fixed `users`, `roles` and `organizations` V1 form
  descriptors are projected by the existing system-layout orchestration edge.
- Permission decisions remain in frozen IAM `SystemLayoutAccessRuntime`; this
  module does not mint permissions or trust tenant headers.
- Frontend and MCP presentation consumers receive Forms presentation documents;
  layout storage never copies IAM records or public field names.

## Data and migration

- Scope: tenant schema only.
- Forward paired versions: `20260802232100..20260802232102`.
- Existing v1/v2 definition bytes and predecessor revisions remain immutable.
- The repair advances only a deterministic installation-created generation-2
  publication. Custom publications are untouched; corrupt deterministic
  predecessors fail closed; replay changes zero rows.
- Rollback is a compatible prior V2 image or a forward fix. There is no reverse
  rewrite, destructive down migration or legacy writer reactivation.

## Risk and verification

- Risk: high for tenant data and compatibility presentation, medium for UI
  parity, low for IAM authorization because IAM is read-only here.
- Required proof: paired-byte and collision checks, module unit/vet/fileguard,
  targeted archguard, fresh/adopted PostgreSQL 18.4, two-tenant isolation,
  custom-layout preservation, corrupt-predecessor denial, replay and concurrency.
- Final Gateway runtime and global manifest normalization remain single-owner
  integration work after all modules freeze.

## Open decisions

No module-local decision remains open. Global parity classification and manifest
normalization stay locked until the final integration pass.
