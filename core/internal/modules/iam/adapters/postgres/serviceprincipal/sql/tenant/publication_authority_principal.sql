WITH authority_lock AS MATERIALIZED (
    SELECT pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'harpy/iam/workflow-publication-principal/v1',
            0
        )
    )
),
principal AS (
    SELECT candidate.id, candidate.kind, candidate.state
    FROM iam_service_principals AS candidate
    WHERE candidate.principal_key = 'workflow.publication'
)
SELECT
    COALESCE(principal.id::text, ''),
    COALESCE(principal.kind, ''),
    COALESCE(principal.state, '')
FROM authority_lock
LEFT JOIN principal ON TRUE;
