-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_content_v1(
    p_scope_schema name,p_export_id uuid,p_actor_id uuid,p_resource_kind text,
    p_token_version bigint,p_authorization_version bigint,p_now timestamptz
)
RETURNS TABLE(
    export_id uuid,job_id uuid,resource_kind text,format text,state text,
    processed_rows bigint,plaintext_bytes bigint,created_at timestamptz,
    ready_at timestamptz,expires_at timestamptz,object_key text,
    object_version text,header_sha256 bytea,ciphertext_sha256 bytea,
    ciphertext_bytes bigint,key_id text,wrapped_dek bytea,wrap_nonce bytea,
    attempt_generation bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_export iam_system_record_exports%ROWTYPE;
BEGIN
    IF NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'export',
        p_token_version,p_authorization_version,false
    ) THEN NULL; END IF;
    SELECT request.* INTO v_export
    FROM iam_system_record_exports AS request
    WHERE request.export_id=p_export_id AND request.actor_id=p_actor_id
      AND request.resource_kind=p_resource_kind;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P0002',
            MESSAGE='IAM export is unavailable';
    END IF;
    IF v_export.state='expired' OR
       (v_export.state='ready' AND v_export.expires_at<=p_now) THEN
        RAISE EXCEPTION USING ERRCODE='P6103',
            MESSAGE='IAM export artifact expired';
    END IF;
    IF v_export.state<>'ready' THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='IAM export artifact is not ready';
    END IF;
    RETURN QUERY SELECT v_export.export_id,v_export.job_id,
        v_export.resource_kind,v_export.format,v_export.state,
        v_export.processed_rows,v_export.plaintext_bytes,v_export.created_at,
        v_export.ready_at,v_export.expires_at,v_export.object_key,
        v_export.object_version,v_export.header_sha256,
        v_export.ciphertext_sha256,v_export.ciphertext_bytes,v_export.key_id,
        v_export.wrapped_dek,v_export.wrap_nonce,v_export.attempt_generation;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_content_v1(
    name,uuid,uuid,text,bigint,bigint,timestamptz
) FROM PUBLIC;
