# Catalog AI Workspace Owner Freeze Handoff

Status: **frozen for global integration**.

The public `catalog/api/aiworkspace.Owner` projects active installed menu
targets for list, describe, search, recent and history. Detail includes the
authoritative target kind and UUID. Workspace uses that reference to call the
target owner; Catalog never queries Forms storage.

Requirements are pinned to `ai.workspace.read` and the exact
`catalog/menu/{item_id}` resource. FORM, WORKFLOW, DASHBOARD and ENDPOINT paths
use target UUIDs; GROUP paths use the menu item UUID. History follows the
immutable Catalog menu definition, not a version-local item ID.

Evidence:

- projection, unavailable/error and immutable-history unit tests are green;
- root owner List/Describe/Search/Recent/History runs through a real runtime
  tenant scope in both fresh and `v1_adopted` PostgreSQL 18.4 lifecycle suites;
- runtime reads only Catalog-owned tables already granted by Catalog migrations.

Intentional redesign candidate: legacy workspace collection discovery included
all form configs and `entity_search_index`; this owner exposes only active
installed Catalog targets. Forms-definition history is also not Catalog menu
history. Global parity inventory must classify both explicitly.
