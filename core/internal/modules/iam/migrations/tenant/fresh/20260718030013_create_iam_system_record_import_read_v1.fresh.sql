-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_import_read_v1(
    p_scope_schema name,p_import_id uuid,p_actor_id uuid,p_resource_kind text,
    p_token_version bigint,p_authorization_version bigint
)
RETURNS TABLE(
    import_id uuid,job_id uuid,resource_kind text,state text,
    processed_rows bigint,created_rows bigint,updated_rows bigint,
    skipped_rows bigint,rejected_rows bigint,created_at timestamptz,
    completed_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'import',
        p_token_version,p_authorization_version,false
    ) THEN NULL; END IF;
    RETURN QUERY
    SELECT request.import_id,request.job_id,request.resource_kind,request.state,
        request.processed_rows,request.created_rows,request.updated_rows,
        request.skipped_rows,request.rejected_rows,request.created_at,
        request.completed_at
    FROM iam_system_record_imports AS request
    WHERE request.import_id=p_import_id AND request.actor_id=p_actor_id
      AND request.resource_kind=p_resource_kind;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_import_read_v1(
    name,uuid,uuid,text,bigint,bigint
) FROM PUBLIC;
