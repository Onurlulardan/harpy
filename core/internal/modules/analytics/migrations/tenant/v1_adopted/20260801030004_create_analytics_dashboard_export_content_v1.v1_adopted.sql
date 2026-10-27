-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_content_v1(
    p_scope_schema name, p_actor_id uuid, p_dashboard_id uuid,
    p_export_id uuid, p_observed_at timestamptz
)
RETURNS TABLE (
    actor_id uuid, export_id uuid, dashboard_id uuid,
    dashboard_version_id uuid, token_version bigint,
    authorization_version bigint, authorization_clock bigint,
    dashboard_manifest_sha256 bytea, pin_sha256 bytea, format text,
    parameters_document bytea, pinned_widgets jsonb,
    maximum_rows bigint, maximum_plaintext_bytes bigint,
    expires_at timestamptz, attempt_generation bigint,
    object_key text, object_version text, key_id text, wrapped_dek bytea,
    wrap_nonce bytea, header_sha256 bytea, ciphertext_sha256 bytea,
    plaintext_sha256 bytea, ciphertext_bytes bigint,
    plaintext_bytes bigint, row_count bigint, content_type text,
    filename text, ready_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_actor_id IS NULL OR
       p_actor_id='00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.uuid_extract_version(p_dashboard_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export content input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    RETURN QUERY
    SELECT item.actor_id,item.export_id,item.dashboard_id,
           item.dashboard_version_id,item.token_version,
           item.authorization_version,item.authorization_clock,
           item.dashboard_manifest_sha256,item.pin_sha256,item.format,
           item.parameters_document,item.pinned_widgets,
           item.maximum_rows,item.maximum_plaintext_bytes,item.expires_at,
           item.attempt_generation,item.object_key,item.object_version,
           item.key_id,item.wrapped_dek,item.wrap_nonce,item.header_sha256,
           item.ciphertext_sha256,item.plaintext_sha256,
           item.ciphertext_bytes,item.plaintext_bytes,item.row_count,
           item.content_type,item.filename,item.ready_at
    FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id
      AND item.dashboard_id=p_dashboard_id
      AND item.state='ready'
      AND item.ready_at IS NOT NULL
      AND item.expires_at>p_observed_at
      AND analytics_dashboard_export_access_v1(
          p_actor_id,item.dashboard_id,item.dashboard_version_id);
    IF NOT FOUND THEN
        IF EXISTS (
            SELECT 1 FROM analytics_dashboard_exports AS item
            WHERE item.export_id=p_export_id
              AND item.dashboard_id=p_dashboard_id
              AND analytics_dashboard_export_access_v1(
                  p_actor_id,item.dashboard_id,item.dashboard_version_id)
        ) THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='analytics dashboard export content is not ready';
        END IF;
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics dashboard export was not found';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_content_v1(
    name,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;

