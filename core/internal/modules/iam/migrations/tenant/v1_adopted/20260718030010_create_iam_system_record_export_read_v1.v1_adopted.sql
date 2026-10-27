-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_read_v1(
    p_scope_schema name,p_export_id uuid,p_actor_id uuid,p_resource_kind text,
    p_token_version bigint,p_authorization_version bigint
)
RETURNS TABLE(
    export_id uuid,job_id uuid,resource_kind text,format text,state text,
    processed_rows bigint,plaintext_bytes bigint,created_at timestamptz,
    ready_at timestamptz,expires_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'export',
        p_token_version,p_authorization_version,false
    ) THEN NULL; END IF;
    RETURN QUERY
    SELECT request.export_id,request.job_id,request.resource_kind,request.format,
        request.state,request.processed_rows,request.plaintext_bytes,
        request.created_at,request.ready_at,request.expires_at
    FROM iam_system_record_exports AS request
    WHERE request.export_id=p_export_id AND request.actor_id=p_actor_id
      AND request.resource_kind=p_resource_kind;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_read_v1(
    name,uuid,uuid,text,bigint,bigint
) FROM PUBLIC;

