-- +goose Up
CREATE VIEW platform.billing_legacy_identity_alias_v1
WITH (security_barrier = true)
AS
WITH ranked_runs AS (
    SELECT run.operation_id, run.run_id, run.authority_id,
           run.scope_kind, run.scope_id,
           pg_catalog.row_number() OVER (
               PARTITION BY run.authority_id
               ORDER BY run.applied_at DESC NULLS LAST,
                        run.updated_at DESC,
                        run.operation_id DESC,
                        run.run_id DESC
           ) AS position
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.state IN ('applied', 'verified', 'rollback_rehearsed')
      AND run.eligible
)
SELECT run.authority_id, run.scope_kind, run.scope_id,
       mapping.record_kind,
       mapping.source_key::uuid AS legacy_id,
       mapping.target_id AS canonical_id
FROM ranked_runs AS run
JOIN platform.billing_v1_adoption_mappings AS mapping
  ON mapping.operation_id = run.operation_id
 AND mapping.run_id = run.run_id
 AND mapping.authority_id = run.authority_id
 AND mapping.scope_kind = run.scope_kind
 AND mapping.scope_id = run.scope_id
WHERE run.position = 1
  AND mapping.record_kind IN (
      'offering', 'discount', 'campaign', 'subscription'
  )
  AND SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'forgecrud_runtime';

GRANT SELECT ON platform.billing_legacy_identity_alias_v1
TO forgecrud_runtime;
