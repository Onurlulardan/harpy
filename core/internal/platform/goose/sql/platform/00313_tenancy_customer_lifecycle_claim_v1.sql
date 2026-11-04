-- +goose Up
CREATE FUNCTION platform.tenancy_customer_lifecycle_claim_v1(
    in_worker_id uuid,
    in_now timestamptz,
    in_seconds integer
)
RETURNS TABLE (
    operation_id uuid,
    control_tenant_id uuid,
    target_tenant_id uuid,
    kind text,
    state text,
    result_revision bigint,
    "false" boolean,
    created_at timestamptz,
    updated_at timestamptz,
    completed_at timestamptz,
    failure_code text,
    schema_name text,
    retention_receipt_id uuid,
    backup_receipt_id uuid,
    contributor_receipt_id uuid,
    evidence_sha256 bytea,
    proof_verified_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        0::bigint,
        false,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        NULL::timestamptz,
        NULL::text,
        ''::text,
        NULL::uuid,
        NULL::uuid,
        NULL::uuid,
        ''::bytea,
        NULL::timestamptz
    ;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_claim_v1(uuid, timestamptz, integer) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_claim_v1(uuid, timestamptz, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_customer_lifecycle_claim_v1(uuid, timestamptz, integer);
