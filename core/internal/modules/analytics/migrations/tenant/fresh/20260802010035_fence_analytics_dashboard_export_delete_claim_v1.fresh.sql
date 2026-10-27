-- +goose Up
ALTER FUNCTION analytics_dashboard_export_delete_claim_v1(
    name,uuid,uuid,uuid,timestamptz
) RENAME TO analytics_dashboard_export_delete_claim_unfenced_v1;
REVOKE ALL ON FUNCTION analytics_dashboard_export_delete_claim_unfenced_v1(
    name,uuid,uuid,uuid,timestamptz
) FROM PUBLIC,forgecrud_runtime;

-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_delete_claim_v1(
    p_scope_schema name,p_export_id uuid,p_delete_job_id uuid,
    p_lease_owner uuid,p_observed_at timestamptz
)
RETURNS TABLE (
    disposition text,state text,actor_id uuid,export_id uuid,
    dashboard_id uuid,dashboard_version_id uuid,
    token_version bigint,authorization_version bigint,
    authorization_clock bigint,dashboard_manifest_sha256 bytea,
    pin_sha256 bytea,format text,parameters_document bytea,
    pinned_widgets jsonb,maximum_rows bigint,
    maximum_plaintext_bytes bigint,expires_at timestamptz,
    attempt_generation bigint,has_artifact boolean,object_key text,
    object_version text,key_id text,wrapped_dek bytea,wrap_nonce bytea,
    header_sha256 bytea,ciphertext_sha256 bytea,plaintext_sha256 bytea,
    ciphertext_bytes bigint,plaintext_bytes bigint,row_count bigint,
    content_type text,filename text,delete_actor_id uuid,delete_job_id uuid,
    delete_attempt_generation bigint,delete_lease_owner uuid
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_platform_kind text;
    v_export analytics_dashboard_exports%ROWTYPE;
    v_expected_kind text;
BEGIN
    IF pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_delete_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_lease_owner) IS DISTINCT FROM 5 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export delete claim input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row
    WHERE identity_row.singleton
      AND identity_row.schema_name=pg_catalog.current_schema();
    v_platform_kind:=
        platform.platform_async_analytics_dashboard_export_maintenance_effect_owner_v1(
            v_tenant_id,p_export_id,p_delete_job_id,p_lease_owner,p_observed_at);
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=v_tenant_id
      AND (item.delete_job_id=p_delete_job_id OR item.expiry_job_id=p_delete_job_id)
    FOR UPDATE;
    v_expected_kind:=CASE WHEN v_export.expiry_job_id=p_delete_job_id THEN
        'analytics.dashboard.export.expire' ELSE
        'analytics.dashboard.export.delete' END;
    IF v_export.export_id IS NULL OR
       v_platform_kind IS DISTINCT FROM v_expected_kind OR
       v_export.maintenance_unknown_job_id IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export delete effect owner changed';
    END IF;
    RETURN QUERY SELECT * FROM analytics_dashboard_export_delete_claim_unfenced_v1(
        p_scope_schema,p_export_id,p_delete_job_id,p_lease_owner,p_observed_at);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_delete_claim_v1(
    name,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_delete_claim_v1(
    name,uuid,uuid,uuid,timestamptz
) TO forgecrud_runtime;

