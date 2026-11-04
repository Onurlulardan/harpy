-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_dashboard_export_content_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 uuid,
    arg_5 timestamptz
)
RETURNS TABLE (
    actor_id uuid,
    export_id uuid,
    dashboard_id uuid,
    version_id uuid,
    token_version bigint,
    authorization_version bigint,
    authorization_clock bigint,
    manifest bytea,
    pin_sha256 bytea,
    format text,
    parameters bytea,
    widgets bytea,
    maximum_rows bigint,
    maximum_bytes bigint,
    expires_at timestamptz,
    generation bigint,
    object_key text,
    object_version text,
    key_id text,
    wrapped_dek bytea,
    wrap_nonce bytea,
    header_sha256 bytea,
    ciphertext_sha256 bytea,
    plaintext_sha256 bytea,
    ciphertext_bytes bigint,
    plaintext_bytes bigint,
    rows bigint,
    content_type text,
    filename text,
    ready_at timestamptz
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
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        0::bigint,
        NULL::bytea,
        NULL::bytea,
        ''::text,
        NULL::bytea,
        NULL::bytea,
        0::bigint,
        0::bigint,
        pg_catalog.clock_timestamp(),
        0::bigint,
        ''::text,
        ''::text,
        ''::text,
        NULL::bytea,
        NULL::bytea,
        NULL::bytea,
        NULL::bytea,
        NULL::bytea,
        0::bigint,
        0::bigint,
        0::bigint,
        ''::text,
        ''::text,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION analytics_dashboard_export_content_v1(name, uuid, uuid, uuid, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_content_v1(name, uuid, uuid, uuid, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_content_v1(name, uuid, uuid, uuid, timestamptz);
