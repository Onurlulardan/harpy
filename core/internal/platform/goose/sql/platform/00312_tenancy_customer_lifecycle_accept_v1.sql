-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_lifecycle_accept_v1(
    in_environment_id uuid,
    in_control_tenant_id uuid,
    in_actor_id uuid,
    in_target_tenant_id uuid,
    in_kind text,
    in_expected_revision bigint,
    in_idempotency_key text,
    in_request_sha256 bytea,
    in_retention_receipt_id uuid,
    in_backup_receipt_id uuid,
    in_contributor_receipt_id uuid,
    in_evidence_sha256 bytea,
    in_verified_at timestamptz,
    in_now timestamptz
)
RETURNS TABLE (
    outcome text,
    operation_id uuid,
    control_tenant_id uuid,
    target_tenant_id uuid,
    kind text,
    state text,
    result_revision bigint,
    replay boolean,
    created_at timestamptz,
    updated_at timestamptz,
    completed_at timestamptz,
    failure_code text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        NULL::uuid,
        in_control_tenant_id,
        in_target_tenant_id,
        in_kind,
        NULL::text,
        NULL::bigint,
        NULL::boolean,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::text
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_accept_v1(uuid, uuid, uuid, uuid, text, bigint, text, bytea, uuid, uuid, uuid, bytea, timestamptz, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_accept_v1(uuid, uuid, uuid, uuid, text, bigint, text, bytea, uuid, uuid, uuid, bytea, timestamptz, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_customer_lifecycle_accept_v1(uuid, uuid, uuid, uuid, text, bigint, text, bytea, uuid, uuid, uuid, bytea, timestamptz, timestamptz);
