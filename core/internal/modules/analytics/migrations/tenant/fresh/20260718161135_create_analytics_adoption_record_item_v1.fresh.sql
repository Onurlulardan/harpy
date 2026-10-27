-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_adoption_record_item_v1(
    p_scope_schema name, p_authority_id uuid, p_tenant_id uuid,
    p_run_id uuid, p_legacy_endpoint_id uuid, p_source_sha256 bytea,
    p_disposition text, p_reason_code text, p_query_id uuid,
    p_query_version_id uuid, p_publication_release_id uuid,
    p_target_sha256 bytea, p_created_at timestamptz
)
RETURNS TABLE (
    run_id uuid, legacy_endpoint_id uuid, source_sha256 bytea,
    disposition text, reason_code text, query_id uuid,
    query_version_id uuid, publication_release_id uuid,
    target_sha256 bytea, created_at timestamptz, replay boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_inserted boolean;
    v_schema text;
BEGIN
    v_schema := pg_catalog.current_schema()::text;
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS NULL OR p_scope_schema::text IS DISTINCT FROM v_schema THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics adoption authority is invalid';
    END IF;
    PERFORM platform.analytics_assert_tenant_authority_v1(
        p_authority_id, p_tenant_id
    );
    PERFORM 1 FROM tenancy_schema_identity AS identity_row
    WHERE identity_row.singleton AND identity_row.tenant_id = p_tenant_id
      AND identity_row.schema_name = v_schema;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics adoption TenantScope is invalid';
    END IF;
    INSERT INTO analytics_v1_adoption_items (
        run_id, legacy_endpoint_id, source_sha256, disposition,
        reason_code, query_id, query_version_id, publication_release_id,
        target_sha256, created_at
    ) VALUES (
        p_run_id, p_legacy_endpoint_id, p_source_sha256, p_disposition,
        p_reason_code, p_query_id, p_query_version_id, p_publication_release_id,
        p_target_sha256, p_created_at
    ) ON CONFLICT DO NOTHING RETURNING true INTO v_inserted;
    IF NOT COALESCE(v_inserted, false) AND NOT EXISTS (
        SELECT 1 FROM analytics_v1_adoption_items AS item
        WHERE item.run_id = p_run_id
          AND item.legacy_endpoint_id = p_legacy_endpoint_id
          AND item.source_sha256 IS NOT DISTINCT FROM p_source_sha256
          AND item.disposition IS NOT DISTINCT FROM p_disposition
          AND item.reason_code IS NOT DISTINCT FROM p_reason_code
          AND item.query_id IS NOT DISTINCT FROM p_query_id
          AND item.query_version_id IS NOT DISTINCT FROM p_query_version_id
          AND item.publication_release_id IS NOT DISTINCT FROM p_publication_release_id
          AND item.target_sha256 IS NOT DISTINCT FROM p_target_sha256
          AND item.created_at IS NOT DISTINCT FROM p_created_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6103',
            MESSAGE = 'analytics adoption replay differs';
    END IF;
    RETURN QUERY
    SELECT item.run_id, item.legacy_endpoint_id, item.source_sha256,
           item.disposition, item.reason_code, item.query_id,
           item.query_version_id, item.publication_release_id,
           item.target_sha256, item.created_at, NOT COALESCE(v_inserted, false)
    FROM analytics_v1_adoption_items AS item
    WHERE item.run_id = p_run_id AND item.legacy_endpoint_id = p_legacy_endpoint_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_adoption_record_item_v1(
    name,uuid,uuid,uuid,uuid,bytea,text,text,uuid,uuid,uuid,bytea,timestamptz
) FROM PUBLIC;
