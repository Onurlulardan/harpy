-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_dashboard_export_platform_claim_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 uuid,
    arg_5 bigint,
    arg_6 uuid
)
RETURNS TABLE (
    claim_state text,
    attempt_id uuid,
    claimed_at timestamptz,
    lease_until timestamptz,
    disposition text,
    state text,
    actor_id uuid,
    export_id uuid,
    dashboard_id uuid,
    version_id uuid,
    job_id uuid,
    token_version bigint,
    authorization_version bigint,
    authorization_clock bigint,
    manifest bytea,
    pin_sha bytea,
    format text,
    parameters bytea,
    widgets bytea,
    maximum_rows bigint,
    maximum_bytes bigint,
    expires_at timestamptz,
    generation bigint,
    lease_owner uuid,
    effect_started boolean,
    object_key text,
    key_id text,
    wrapped_dek bytea,
    wrap_nonce bytea,
    header_sha bytea,
    ciphertext_sha bytea,
    plaintext_sha bytea,
    ciphertext_bytes bigint,
    plaintext_bytes bigint,
    rows bigint,
    content_type text,
    filename text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        NULL::text,
        NULL::uuid,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::text,
        NULL::text,
        NULL::uuid,
        NULL::uuid,
        NULL::uuid,
        NULL::uuid,
        NULL::uuid,
        NULL::bigint,
        NULL::bigint,
        NULL::bigint,
        NULL::bytea,
        NULL::bytea,
        NULL::text,
        NULL::bytea,
        NULL::bytea,
        NULL::bigint,
        NULL::bigint,
        NULL::timestamptz,
        NULL::bigint,
        NULL::uuid,
        NULL::boolean,
        NULL::text,
        NULL::text,
        NULL::bytea,
        NULL::bytea,
        NULL::bytea,
        NULL::bytea,
        NULL::bytea,
        NULL::bigint,
        NULL::bigint,
        NULL::bigint,
        NULL::text,
        NULL::text;
$$;

REVOKE ALL ON FUNCTION analytics_dashboard_export_platform_claim_v1(name, uuid, uuid, uuid, bigint, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_platform_claim_v1(name, uuid, uuid, uuid, bigint, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_platform_claim_v1(name, uuid, uuid, uuid, bigint, uuid);
