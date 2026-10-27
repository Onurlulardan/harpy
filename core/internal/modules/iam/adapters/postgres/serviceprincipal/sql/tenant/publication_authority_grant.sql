WITH authority_lock AS MATERIALIZED (
    SELECT pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'harpy/iam/workflow-publication-grant/v1/' || $1::text,
            0
        )
    )
),
authority_grant AS (
    SELECT
        head.id,
        head.principal_id,
        head.current_revision,
        revision.capability_key,
        head.state,
        revision.scope_kind,
        revision.owner_domain,
        revision.resource_kind,
        revision.resource_id,
        revision.manifest_sha256
    FROM iam_capability_grants AS head
    JOIN iam_capability_grant_revisions AS revision
      ON revision.grant_id = head.id
     AND revision.grant_revision = head.current_revision
    WHERE head.id = $1
)
SELECT
    COALESCE(authority_grant.id::text, ''),
    COALESCE(authority_grant.principal_id::text, ''),
    COALESCE(authority_grant.current_revision, 0),
    COALESCE(authority_grant.capability_key, ''),
    COALESCE(authority_grant.state, ''),
    COALESCE(authority_grant.scope_kind, ''),
    COALESCE(authority_grant.owner_domain, ''),
    COALESCE(authority_grant.resource_kind, ''),
    COALESCE(authority_grant.resource_id::text, ''),
    COALESCE(pg_catalog.encode(authority_grant.manifest_sha256, 'hex'), '')
FROM authority_lock
LEFT JOIN authority_grant ON TRUE;
