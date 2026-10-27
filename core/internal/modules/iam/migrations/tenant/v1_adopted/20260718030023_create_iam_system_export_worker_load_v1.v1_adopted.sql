-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_worker_load_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_resource_kind text
)
RETURNS TABLE(
    actor_id uuid,token_version bigint,authorization_version bigint,
    catalog_version_id uuid,catalog_sha256 bytea,query_document bytea,
    query_sha256 bytea,format text,state text,attempt_generation bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
BEGIN
    SELECT identity.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity
    WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema();
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       v_tenant_id IS DISTINCT FROM p_tenant_id OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       p_resource_kind NOT IN ('users','roles','organizations') THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM export worker load scope is invalid';
    END IF;
    RETURN QUERY
    SELECT request.actor_id,request.token_version,
        request.authorization_version,request.catalog_version_id,
        request.catalog_sha256,request.query_document,request.query_sha256,
        request.format,request.state,request.attempt_generation
    FROM iam_system_record_exports AS request
    WHERE request.export_id=p_export_id AND request.tenant_id=p_tenant_id
      AND request.job_id=p_job_id AND request.resource_kind=p_resource_kind;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6044',
            MESSAGE='IAM export worker request was not found';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_worker_load_v1(
    name,uuid,uuid,uuid,text
) FROM PUBLIC;
