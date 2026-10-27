-- +goose Up
CREATE VIEW iam_rest_query_organizations_v3
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
    ), 'NFC') AS slug_folded,
    COALESCE((
        SELECT pg_catalog.array_agg(membership.user_id ORDER BY membership.user_id)
        FROM iam_organization_memberships AS membership
        WHERE membership.organization_id = organization.id
          AND membership.membership_kind = 'owner'
          AND membership.state = 'active'
    ), ARRAY[]::uuid[]) AS owner_ids
FROM iam_organizations AS organization
WHERE SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'harpy_runtime';

GRANT SELECT ON iam_rest_query_organizations_v3 TO harpy_runtime;
