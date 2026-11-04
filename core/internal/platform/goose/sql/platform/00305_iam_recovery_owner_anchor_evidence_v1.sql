-- +goose Up
CREATE FUNCTION platform.iam_recovery_owner_anchor_evidence_v1(
    in_environment_id text
)
RETURNS TABLE (
    environment_idtext text,
    tenant_idtext text,
    owner_user_idtext text,
    owner_binding_idtext text,
    owner_binding_generation bigint,
    anchor_idtext text,
    anchor_generation bigint,
    credential_id bytea,
    public_key_cose bytea,
    aaguidtext text,
    attestation_format text,
    attestation_type text,
    rp_id text,
    allowed_origins jsonb,
    allowed_origins_sha256 bytea,
    user_verification text,
    backup_eligible boolean,
    backup_state boolean,
    sign_count bigint,
    counter_policy text,
    protected_head_generation bigint,
    protected_state_sha256 bytea,
    protected_ack_sha256 bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        0::bigint,
        ''::bytea,
        ''::bytea,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        '{}'::jsonb,
        ''::bytea,
        ''::text,
        false,
        false,
        0::bigint,
        ''::text,
        0::bigint,
        ''::bytea,
        ''::bytea
    ;
$$;

REVOKE ALL ON FUNCTION platform.iam_recovery_owner_anchor_evidence_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_recovery_owner_anchor_evidence_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_recovery_owner_anchor_evidence_v1(text);
