-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_anchor_stage_registration_v1(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_actor_id uuid,
    in_idempotency_key text,
    in_request_sha256 bytea,
    in_credential_id bytea,
    in_public_key_cose bytea,
    in_aaguid uuid,
    in_attestation_format text,
    in_attestation_type text,
    in_rpid text,
    in_origins jsonb,
    in_origins_sha256 bytea,
    in_sign_count bigint,
    in_occurred_at timestamptz
)
RETURNS TABLE (
    anchor_id bytea,
    anchor_set_sha256 bytea,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::bytea,
        ''::bytea,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_stage_registration_v1(uuid, uuid, uuid, uuid, text, bytea, bytea, bytea, uuid, text, text, text, jsonb, bytea, bigint, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_stage_registration_v1(uuid, uuid, uuid, uuid, text, bytea, bytea, bytea, uuid, text, text, text, jsonb, bytea, bigint, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_recovery_anchor_stage_registration_v1(uuid, uuid, uuid, uuid, text, bytea, bytea, bytea, uuid, text, text, text, jsonb, bytea, bigint, timestamptz);
