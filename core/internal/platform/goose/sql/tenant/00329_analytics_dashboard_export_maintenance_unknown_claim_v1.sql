-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_dashboard_export_maintenance_unknown_claim_v1(
    arg_1 name,
    arg_2 text,
    arg_3 uuid,
    arg_4 name,
    arg_5 uuid,
    arg_6 text,
    arg_7 bigint,
    arg_8 bigint,
    arg_9 uuid,
    arg_10 uuid,
    arg_11 bigint,
    arg_12 bigint,
    arg_13 uuid,
    arg_14 uuid,
    arg_15 bytea,
    arg_16 timestamptz,
    arg_17 uuid,
    arg_18 timestamptz
)
RETURNS TABLE (
    disposition text,
    generation bigint,
    lease_until timestamptz,
    mode text,
    state text,
    format text,
    object_key text,
    object_version text,
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
    filename text,
    render_gen bigint,
    delete_attempt_generation bigint,
    local_state_sha bytea,
    local_state_at timestamptz,
    delete_provider_id text,
    delete_response_sha bytea,
    delete_absence_sha bytea,
    delete_verified_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        NULL::bigint,
        NULL::timestamptz,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::text,
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
        NULL::text,
        NULL::bigint,
        NULL::bigint,
        NULL::bytea,
        NULL::timestamptz,
        NULL::text,
        NULL::bytea,
        NULL::bytea,
        NULL::timestamptz;
$$;

REVOKE ALL ON FUNCTION analytics_dashboard_export_maintenance_unknown_claim_v1(name, text, uuid, name, uuid, text, bigint, bigint, uuid, uuid, bigint, bigint, uuid, uuid, bytea, timestamptz, uuid, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_maintenance_unknown_claim_v1(name, text, uuid, name, uuid, text, bigint, bigint, uuid, uuid, bigint, bigint, uuid, uuid, bytea, timestamptz, uuid, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_maintenance_unknown_claim_v1(name, text, uuid, name, uuid, text, bigint, bigint, uuid, uuid, bigint, bigint, uuid, uuid, bytea, timestamptz, uuid, timestamptz);
