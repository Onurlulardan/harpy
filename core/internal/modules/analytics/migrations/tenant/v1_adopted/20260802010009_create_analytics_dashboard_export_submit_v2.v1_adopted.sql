-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_submit_v2(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea,
    p_export_id uuid, p_job_id uuid, p_expiry_job_id uuid,
    p_dashboard_id uuid, p_format text, p_parameters bytea,
    p_token_version bigint, p_authorization_version bigint,
    p_authorization_clock bigint, p_maximum_rows bigint,
    p_maximum_plaintext_bytes bigint, p_occurred_at timestamptz,
    p_expires_at timestamptz
)
RETURNS TABLE (
    export_id uuid, dashboard_id uuid, dashboard_version_id uuid, job_id uuid,
    effective_state text, stored_state text, format text, filename text,
    content_type text, failure_code text, row_count bigint,
    plaintext_bytes bigint, cancellation_requested boolean,
    created_at timestamptz, updated_at timestamptz, expires_at timestamptz,
    replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_export_id uuid;
    v_replayed boolean;
    v_existing_expiry_job_id uuid;
BEGIN
    IF pg_catalog.uuid_extract_version(p_expiry_job_id) IS DISTINCT FROM 7 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export expiry job is invalid';
    END IF;
    SELECT submitted.export_id,submitted.replayed
    INTO v_export_id,v_replayed
    FROM analytics_dashboard_export_submit_v1(
        p_scope_schema,p_operation_id,p_actor_id,p_idempotency_key,
        p_request_sha256,p_export_id,p_job_id,p_dashboard_id,p_format,
        p_parameters,p_token_version,p_authorization_version,
        p_authorization_clock,p_maximum_rows,p_maximum_plaintext_bytes,
        p_occurred_at,p_expires_at
    ) AS submitted;
    SELECT item.expiry_job_id INTO v_existing_expiry_job_id
    FROM analytics_dashboard_exports AS item
    WHERE item.export_id=v_export_id FOR UPDATE;
    IF v_replayed THEN
        IF v_existing_expiry_job_id IS NULL THEN
            RAISE EXCEPTION USING ERRCODE='P6105',
                MESSAGE='analytics dashboard export expiry schedule is missing';
        END IF;
    ELSE
        UPDATE analytics_dashboard_exports
        SET expiry_job_id=p_expiry_job_id,row_revision=row_revision+1,
            updated_at=p_occurred_at
        WHERE analytics_dashboard_exports.export_id=v_export_id
          AND expiry_job_id IS NULL;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='40001',
                MESSAGE='analytics dashboard export expiry schedule changed';
        END IF;
    END IF;
    RETURN QUERY
    SELECT snapshot.*,v_replayed
    FROM analytics_dashboard_export_snapshot_v1(
        p_scope_schema,p_actor_id,p_dashboard_id,v_export_id,p_occurred_at
    ) AS snapshot;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_submit_v2(
    name,uuid,uuid,text,bytea,uuid,uuid,uuid,uuid,text,bytea,
    bigint,bigint,bigint,bigint,bigint,timestamptz,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_submit_v2(
    name,uuid,uuid,text,bytea,uuid,uuid,uuid,uuid,text,bytea,
    bigint,bigint,bigint,bigint,bigint,timestamptz,timestamptz
) TO harpy_runtime;

