-- +goose Up
WITH eligible AS (
    SELECT role.id AS role_id, role.is_protected,
           permission.id AS permission_id
    FROM iam_roles AS role
    JOIN iam_permissions AS permission ON permission.match_kind = 'exact'
     AND permission.state = 'active' AND permission.is_protected
    JOIN iam_resources AS resource ON resource.id = permission.resource_id
     AND resource.owner_domain = 'work'
     AND resource.resource_kind = 'system_resource'
     AND resource.is_protected AND resource.state = 'active'
     AND resource.resource_key LIKE 'work.%'
    JOIN iam_actions AS action ON action.id = permission.action_id
     AND action.state = 'active' AND NOT action.is_protected
    WHERE role.status = 'active'
      AND NOT role.is_protected
      AND role.name IN ('Tenant Yöneticisi', 'Yönetici')
)
UPDATE iam_role_permission_bindings AS binding
SET state = 'active', revision = binding.revision + 1,
    updated_at = pg_catalog.transaction_timestamp(),
    activated_at = pg_catalog.transaction_timestamp(), revoked_at = NULL
FROM eligible
WHERE binding.role_id = eligible.role_id
  AND binding.permission_id = eligible.permission_id
  AND binding.state = 'revoked';

WITH eligible AS (
    SELECT role.id AS role_id, role.is_protected,
           permission.id AS permission_id
    FROM iam_roles AS role
    JOIN iam_permissions AS permission ON permission.match_kind = 'exact'
     AND permission.state = 'active' AND permission.is_protected
    JOIN iam_resources AS resource ON resource.id = permission.resource_id
     AND resource.owner_domain = 'work'
     AND resource.resource_kind = 'system_resource'
     AND resource.is_protected AND resource.state = 'active'
     AND resource.resource_key LIKE 'work.%'
    JOIN iam_actions AS action ON action.id = permission.action_id
     AND action.state = 'active' AND NOT action.is_protected
    WHERE role.status = 'active'
      AND NOT role.is_protected
      AND role.name IN ('Tenant Yöneticisi', 'Yönetici')
)
INSERT INTO iam_role_permission_bindings (
    id, role_id, role_is_protected, permission_id,
    permission_match_kind, state, revision, created_at, updated_at,
    activated_at, revoked_at
)
SELECT pg_catalog.uuidv7(), eligible.role_id, eligible.is_protected,
       eligible.permission_id, 'exact', 'active', 1,
       pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp(),
       pg_catalog.transaction_timestamp(), NULL
FROM eligible
ON CONFLICT (role_id, permission_id) DO NOTHING;

WITH bumped AS (
    UPDATE iam_authorization_clock
    SET clock_value = clock_value + 1, revision = revision + 1,
        updated_at = pg_catalog.transaction_timestamp()
    WHERE singleton
    RETURNING clock_value
)
SELECT iam_rbac_bump_all_users_v1(clock_value)
FROM bumped;

SELECT 1 / CASE WHEN NOT EXISTS (
    SELECT 1
    FROM iam_roles AS role
    JOIN iam_permissions AS permission ON permission.match_kind = 'exact'
     AND permission.state = 'active' AND permission.is_protected
    JOIN iam_resources AS resource ON resource.id = permission.resource_id
     AND resource.owner_domain = 'work'
     AND resource.resource_kind = 'system_resource'
     AND resource.is_protected AND resource.state = 'active'
     AND resource.resource_key LIKE 'work.%'
    WHERE role.status = 'active' AND NOT role.is_protected
      AND role.name IN ('Tenant Yöneticisi', 'Yönetici')
      AND NOT EXISTS (
          SELECT 1
          FROM iam_role_permission_bindings AS binding
          WHERE binding.role_id = role.id
            AND binding.permission_id = permission.id
            AND binding.permission_match_kind = 'exact'
            AND binding.state = 'active'
      )
) THEN 1 ELSE 0 END AS iam_work_access_grant_assertion;
