-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_accept_v2(
    in_operation_id uuid,
    in_environment_id uuid,
    in_kind text,
    in_idempotency_key text,
    in_request_sha256 text,
    in_manifest_sha256 text,
    in_step_plan_sha256 text,
    in_deployment_mode text,
    in_tenant_id uuid,
    in_schema_name name,
    in_binding_purpose text,
    in_first_user_id uuid,
    in_root_organization_id uuid,
    in_protected_admin_binding_id uuid,
    in_instance_owner_binding_id uuid,
    in_offering_version_id uuid,
    in_offering_revision bigint,
    in_offering_sha256 text,
    in_offering_effective_at timestamptz,
    in_source_sequence bigint,
    in_source_sha256 text,
    in_recovery_generation bigint,
    in_trust_bundle_sha256 text,
    in_registry_binding_id uuid,
    in_scope_kind text,
    in_scope_id uuid,
    in_authority_generation bigint,
    in_previous_head_sha256 text,
    in_state_sha256 text,
    in_proposal_sha256 text,
    in_created_at timestamptz
)
RETURNS TABLE (
    operation_id uuid,
    state text,
    revision bigint,
    updated_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    WITH clock AS (
        SELECT CASE
            WHEN pg_catalog.clock_timestamp() >= in_created_at THEN pg_catalog.clock_timestamp()
            ELSE in_created_at
        END AS updated_at
    ), current_row AS (
        SELECT operation.state, operation.revision
        FROM platform.tenancy_onboarding_operations AS operation
        WHERE operation.operation_id = in_operation_id
    ), nxt AS (
        SELECT
            CASE
                WHEN current_row.state IS NULL THEN 'accepted'
                WHEN current_row.state = 'accepted' AND in_deployment_mode = 'saas'
                    THEN 'authority_pending'
                WHEN current_row.state = 'accepted' THEN 'migration_ready'
                WHEN current_row.state = 'migration_ready' THEN 'succeeded'
                WHEN current_row.state = 'authority_pending' THEN 'authority_acknowledged'
                WHEN current_row.state = 'authority_acknowledged' THEN 'succeeded'
                ELSE current_row.state
            END AS state,
            CASE
                WHEN current_row.state IS NULL THEN 1
                WHEN current_row.state = 'succeeded' THEN current_row.revision
                ELSE current_row.revision + 1
            END AS revision
        FROM clock
        LEFT JOIN current_row ON true
    ), stored AS (
        INSERT INTO platform.tenancy_onboarding_operations (
            operation_id, tenant_id, first_owner_user_id, kind, step_plan_sha256,
            deployment_mode, state, revision, attempt_number, updated_at
        )
        SELECT in_operation_id, in_tenant_id, in_first_user_id, in_kind,
               in_step_plan_sha256, in_deployment_mode, nxt.state, nxt.revision,
               0, clock.updated_at
        FROM nxt
        CROSS JOIN clock
        ON CONFLICT (operation_id) DO UPDATE SET
            state = EXCLUDED.state,
            revision = EXCLUDED.revision,
            updated_at = EXCLUDED.updated_at
        RETURNING operation_id, state, revision, updated_at
    )
    SELECT stored.operation_id, stored.state, stored.revision, stored.updated_at
    FROM stored;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_accept_v2(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_accept_v2(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_accept_v2(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz);
