-- +goose Up
CREATE FUNCTION platform.platformops_claim_restore_execution_v5(
    in_operation_id text,
    in_attempt_id text,
    in_lease_token text,
    in_plan_sha256 text
)
RETURNS TABLE (
    attempt_id text,
    lease_token text,
    attempt_generation bigint,
    restore_state text,
    operation_id text,
    recovery_set_id text,
    environment_id text,
    target_disposition text,
    plan_sha256 text,
    plan_bytes bytea,
    safety_recovery_set_id text,
    safety_recovery_set_sha256 text,
    waiver_reason text,
    probe_sha256 text,
    destructive_confirmed boolean,
    manifest_sha256 text,
    manifest_bytes bytea,
    previous_database_system_id text,
    component_inventory jsonb,
    control_database_locator_sha256 text,
    scratch_database_locator_sha256 text,
    drain_ack_sha256 text,
    switch_ack_sha256 text,
    archive_readback_sha256 text,
    new_database_system_id text,
    object_merkle_sha256 text,
    locator_remap_sha256 text,
    safety_evidence_sha256 text,
    database_now timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::bytea,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        false,
        ''::text,
        ''::bytea,
        ''::text,
        '{}'::jsonb,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.platformops_claim_restore_execution_v5(text, text, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_claim_restore_execution_v5(text, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_claim_restore_execution_v5(text, text, text, text);
