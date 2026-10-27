-- +goose Up
CREATE VIEW iam_rest_query_users_v2
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
    principal.updated_at,
    pg_catalog.normalize(profile_display.value, 'NFC') AS display_name,
    pg_catalog.normalize(pg_catalog.casefold(
        pg_catalog.normalize(profile_display.value, 'NFC')
            COLLATE pg_catalog.pg_unicode_fast
    ), 'NFC') AS display_name_folded
FROM iam_users AS principal
LEFT JOIN LATERAL (
    SELECT COALESCE(
        NULLIF(principal.profile_values ->> profile_field.id::text, ''),
        NULLIF(principal.profile_values ->> 'display_name', '')
    ) AS value
    FROM iam_profile_fields AS profile_field
    WHERE profile_field.resource_kind = 'user'
      AND profile_field.field_key = 'display_name'
      AND profile_field.value_kind = 'string'
      AND profile_field.cardinality = 'single'
      AND profile_field.classification = 'public'
      AND profile_field.state = 'active'
) AS profile_display ON true
WHERE SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'forgecrud_runtime';

CREATE VIEW iam_rest_query_organizations_v2
AS
SELECT
    organization.id,
    organization.name,
    organization.slug,
    organization.status,
    organization.parent_id,
    COALESCE((
        SELECT pg_catalog.array_agg(membership.user_id ORDER BY membership.user_id)
        FROM iam_organization_memberships AS membership
        WHERE membership.organization_id = organization.id
          AND membership.state = 'active'
    ), ARRAY[]::uuid[]) AS user_ids,
    COALESCE((
        SELECT pg_catalog.array_agg(role.id ORDER BY role.id)
        FROM iam_roles AS role
        WHERE role.organization_id = organization.id AND role.status = 'active'
    ), ARRAY[]::uuid[]) AS role_ids,
    organization.profile_schema_version_id,
    organization.profile_values,
    organization.revision,
    organization.created_at,
    organization.updated_at,
    pg_catalog.normalize(pg_catalog.casefold(
        pg_catalog.normalize(organization.name, 'NFC')
            COLLATE pg_catalog.pg_unicode_fast
    ), 'NFC') AS name_folded,
    pg_catalog.normalize(pg_catalog.casefold(
        pg_catalog.normalize(organization.slug, 'NFC')
            COLLATE pg_catalog.pg_unicode_fast
    ), 'NFC') AS slug_folded
FROM iam_organizations AS organization
WHERE SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'forgecrud_runtime';

CREATE VIEW iam_rest_query_roles_v2
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
    role.updated_at,
    pg_catalog.normalize(pg_catalog.casefold(
        pg_catalog.normalize(role.name, 'NFC')
            COLLATE pg_catalog.pg_unicode_fast
    ), 'NFC') AS name_folded,
    pg_catalog.normalize(pg_catalog.casefold(
        pg_catalog.normalize(role.description, 'NFC')
            COLLATE pg_catalog.pg_unicode_fast
    ), 'NFC') AS description_folded
FROM iam_roles AS role
WHERE SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'forgecrud_runtime';

GRANT SELECT ON iam_rest_query_users_v2 TO forgecrud_runtime;
GRANT SELECT ON iam_rest_query_roles_v2 TO forgecrud_runtime;
GRANT SELECT ON iam_rest_query_organizations_v2 TO forgecrud_runtime;
