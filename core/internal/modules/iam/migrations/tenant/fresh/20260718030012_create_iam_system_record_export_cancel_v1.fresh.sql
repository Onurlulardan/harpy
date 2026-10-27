-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_cancel_v1(
    p_scope_schema name,p_export_id uuid,p_actor_id uuid,p_resource_kind text,
    p_token_version bigint,p_authorization_version bigint,p_now timestamptz
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
DECLARE
    v_export iam_system_record_exports%ROWTYPE;
BEGIN
    IF NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'export',
        p_token_version,p_authorization_version,true
    ) THEN NULL; END IF;
    SELECT request.* INTO v_export
    FROM iam_system_record_exports AS request
    WHERE request.export_id=p_export_id AND request.actor_id=p_actor_id
      AND request.resource_kind=p_resource_kind
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P0002',
            MESSAGE='IAM export is unavailable';
    END IF;
    IF v_export.state IN ('waiting','running') THEN
        UPDATE iam_system_record_exports SET state='cancelled',completed_at=p_now
        WHERE iam_system_record_exports.export_id=p_export_id;
        UPDATE iam_system_transfer_admissions
        SET state='released',tenant_run_slot=NULL,principal_wait_slot=NULL,
            released_at=p_now
        WHERE transfer_kind='export' AND transfer_id=p_export_id
          AND state IN ('waiting','running');
        v_export.state := 'cancelled';
    ELSIF v_export.state<>'cancelled' THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='IAM export cannot be cancelled';
    END IF;
    RETURN QUERY SELECT v_export.export_id,v_export.job_id,
        v_export.resource_kind,v_export.format,v_export.state,
        v_export.processed_rows,v_export.plaintext_bytes,v_export.created_at,
        v_export.ready_at,v_export.expires_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_cancel_v1(
    name,uuid,uuid,text,bigint,bigint,timestamptz
) FROM PUBLIC;
