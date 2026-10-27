-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_import_cancel_v1(
    p_scope_schema name,p_import_id uuid,p_actor_id uuid,p_resource_kind text,
    p_token_version bigint,p_authorization_version bigint,p_now timestamptz
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
DECLARE
    v_import iam_system_record_imports%ROWTYPE;
BEGIN
    IF NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'import',
        p_token_version,p_authorization_version,true
    ) THEN NULL; END IF;
    SELECT request.* INTO v_import
    FROM iam_system_record_imports AS request
    WHERE request.import_id=p_import_id AND request.actor_id=p_actor_id
      AND request.resource_kind=p_resource_kind
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P0002',
            MESSAGE='IAM import is unavailable';
    END IF;
    IF v_import.state IN ('waiting','running') THEN
        UPDATE iam_system_record_imports SET state='cancelled',completed_at=p_now
        WHERE iam_system_record_imports.import_id=p_import_id;
        UPDATE iam_system_transfer_admissions
        SET state='released',tenant_run_slot=NULL,principal_wait_slot=NULL,
            released_at=p_now
        WHERE transfer_kind='import' AND transfer_id=p_import_id
          AND state IN ('waiting','running');
        v_import.state := 'cancelled';
        v_import.completed_at := p_now;
    ELSIF v_import.state<>'cancelled' THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='IAM import cannot be cancelled';
    END IF;
    RETURN QUERY SELECT v_import.import_id,v_import.job_id,
        v_import.resource_kind,v_import.state,v_import.processed_rows,
        v_import.created_rows,v_import.updated_rows,v_import.skipped_rows,
        v_import.rejected_rows,v_import.created_at,v_import.completed_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_import_cancel_v1(
    name,uuid,uuid,text,bigint,bigint,timestamptz
) FROM PUBLIC;
