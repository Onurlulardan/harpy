-- +goose Up
DROP FUNCTION platform.platformops_claim_next_recovery_set_v3(text, text);

-- +goose StatementBegin
CREATE FUNCTION platform.platformops_claim_next_recovery_set_v3(
    in_attempt_id text,
    in_lease_token text
)
RETURNS TABLE (
    attempt_id text,
    lease_token text,
    attempt_generation bigint,
    recovery_state text,
    recovery_set_id text,
    operation_id text,
    environment_id text,
    sequence bigint,
    primary_binding_id text,
    tenant_set_sha256 text,
    owner_binding_id text,
    owner_binding_generation bigint,
    database_lineage_id text,
    database_system_id text,
    recovery_signing_generation bigint,
    recovery_trust_bundle_sha256 text,
    component_inventory jsonb,
    database_now timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        '{}'::jsonb,
        pg_catalog.clock_timestamp()
    WHERE false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_claim_next_recovery_set_v3(text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_claim_next_recovery_set_v3(text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_claim_next_recovery_set_v3(text, text);
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_claim_next_recovery_set_v3(
    in_attempt_id text,
    in_lease_token text
)
RETURNS TABLE (
    attempt_id text,
    lease_token text,
    attempt_generation bigint,
    recovery_state text,
    recovery_set_id text,
    operation_id text,
    environment_id text,
    sequence bigint,
    primary_binding_id text,
    tenant_set_sha256 text,
    owner_binding_id text,
    owner_binding_generation bigint,
    database_lineage_id text,
    database_system_id text,
    recovery_signing_generation bigint,
    recovery_trust_bundle_sha256 text,
    component_inventory jsonb,
    database_now timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        '{}'::jsonb,
        pg_catalog.clock_timestamp();
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platformops_claim_next_recovery_set_v3(text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_claim_next_recovery_set_v3(text, text) TO harpy_api, harpy_worker;
