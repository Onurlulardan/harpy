-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_unknown_cleanup_claim_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 uuid,
    arg_5 uuid,
    arg_6 bigint,
    arg_7 uuid,
    arg_8 uuid,
    arg_9 bytea,
    arg_10 uuid,
    arg_11 timestamptz
)
RETURNS TABLE (
    disposition text,
    generation bigint,
    lease_until timestamptz,
    format text,
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
    filename text,
    effect_started_at timestamptz,
    cancellation_requested boolean,
    expires_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        NULL::bigint,
        NULL::timestamptz,
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
        NULL::timestamptz,
        NULL::boolean,
        NULL::timestamptz;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_unknown_cleanup_claim_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, uuid, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_unknown_cleanup_claim_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, uuid, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_unknown_cleanup_claim_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, uuid, timestamptz);
