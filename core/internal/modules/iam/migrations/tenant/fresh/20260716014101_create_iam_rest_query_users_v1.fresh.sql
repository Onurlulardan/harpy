-- +goose Up
CREATE VIEW iam_rest_query_users_v1
AS
SELECT
    principal.id,
    (SELECT identifier.display_value
     FROM iam_user_identifiers AS identifier
     WHERE identifier.user_id = principal.id
       AND identifier.kind = 'email_v1'
       AND identifier.state = 'active') AS email,
    (SELECT identifier.display_value
     FROM iam_user_identifiers AS identifier
     WHERE identifier.user_id = principal.id
       AND identifier.kind = 'portal_v1'
       AND identifier.state = 'active') AS portal_identifier,
    principal.status,
    principal.identity_kind,
    COALESCE((
        SELECT pg_catalog.array_agg(assignment.role_id ORDER BY assignment.role_id)
        FROM iam_user_role_assignments AS assignment
        WHERE assignment.user_id = principal.id AND assignment.state = 'active'
    ), ARRAY[]::uuid[]) AS role_ids,
    COALESCE((
        SELECT pg_catalog.array_agg(membership.organization_id
                                   ORDER BY membership.organization_id)
        FROM iam_organization_memberships AS membership
        WHERE membership.user_id = principal.id AND membership.state = 'active'
    ), ARRAY[]::uuid[]) AS organization_ids,
    principal.profile_schema_version_id,
    principal.profile_values,
    principal.revision,
    principal.created_at,
    principal.updated_at
FROM iam_users AS principal
WHERE SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'harpy_runtime';
