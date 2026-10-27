-- +goose Up
CREATE VIEW iam_rest_query_roles_v1
AS
SELECT
    role.id,
    role.scope_kind,
    role.scope_id,
    role.organization_id,
    role.name,
    role.description,
    role.status,
    role.is_protected,
    role.is_default,
    COALESCE((
        SELECT pg_catalog.array_agg(assignment.user_id ORDER BY assignment.user_id)
        FROM iam_user_role_assignments AS assignment
        WHERE assignment.role_id = role.id AND assignment.state = 'active'
    ), ARRAY[]::uuid[]) AS user_ids,
    role.profile_schema_version_id,
    role.profile_values,
    role.revision,
    role.created_at,
    role.updated_at
FROM iam_roles AS role
WHERE SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'harpy_runtime';
