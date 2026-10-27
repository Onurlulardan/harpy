-- +goose Up
WITH eligible AS (
    SELECT role.id AS role_id, role.is_protected,
           permission.id AS permission_id
    FROM iam_roles AS role
    CROSS JOIN LATERAL (
        SELECT permission.id, action.action_key
        FROM iam_permissions AS permission
        JOIN iam_resources AS resource ON resource.id = permission.resource_id
        JOIN iam_actions AS action ON action.id = permission.action_id
        WHERE permission.match_kind = 'exact'
          AND permission.state = 'active'
          AND resource.resource_key = 'bug_reports'
          AND resource.owner_domain = 'support'
          AND resource.state = 'active'
          AND action.action_key IN ('create', 'read_own', 'read', 'manage')
          AND action.state = 'active'
    ) AS permission
    WHERE role.status = 'active'
      AND (
          permission.action_key IN ('create', 'read_own')
          OR permission.action_key IN ('read', 'manage')
             AND (
                 role.is_protected
                 OR role.name IN (
                     'Tenant Yöneticisi', 'Yönetici', 'Tenant Admin'
                 )
             )
      )
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
    CROSS JOIN LATERAL (
        SELECT permission.id, action.action_key
        FROM iam_permissions AS permission
        JOIN iam_resources AS resource ON resource.id = permission.resource_id
        JOIN iam_actions AS action ON action.id = permission.action_id
        WHERE permission.match_kind = 'exact'
          AND permission.state = 'active'
          AND resource.resource_key = 'bug_reports'
          AND resource.owner_domain = 'support'
          AND resource.state = 'active'
          AND action.action_key IN ('create', 'read_own', 'read', 'manage')
          AND action.state = 'active'
    ) AS permission
    WHERE role.status = 'active'
      AND (
          permission.action_key IN ('create', 'read_own')
          OR permission.action_key IN ('read', 'manage')
             AND (
                 role.is_protected
                 OR role.name IN (
                     'Tenant Yöneticisi', 'Yönetici', 'Tenant Admin'
                 )
             )
      )
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

SELECT 1 / CASE WHEN EXISTS (
    SELECT 1 FROM iam_roles WHERE status = 'active'
) AND NOT EXISTS (
    SELECT 1
    FROM iam_roles AS role
    CROSS JOIN LATERAL (
        SELECT permission.id, action.action_key
        FROM iam_permissions AS permission
        JOIN iam_resources AS resource ON resource.id = permission.resource_id
        JOIN iam_actions AS action ON action.id = permission.action_id
        WHERE permission.match_kind = 'exact'
          AND permission.state = 'active'
          AND resource.resource_key = 'bug_reports'
          AND resource.owner_domain = 'support'
          AND resource.state = 'active'
          AND action.action_key IN ('create', 'read_own', 'read', 'manage')
          AND action.state = 'active'
    ) AS permission
    WHERE role.status = 'active'
      AND (
          permission.action_key IN ('create', 'read_own')
          OR permission.action_key IN ('read', 'manage')
             AND (
                 role.is_protected
                 OR role.name IN (
                     'Tenant Yöneticisi', 'Yönetici', 'Tenant Admin'
                 )
             )
      )
      AND NOT EXISTS (
          SELECT 1
          FROM iam_role_permission_bindings AS binding
          WHERE binding.role_id = role.id
            AND binding.permission_id = permission.id
            AND binding.permission_match_kind = 'exact'
            AND binding.state = 'active'
      )
) THEN 1 ELSE 0 END AS iam_support_bug_report_grant_assertion;
