-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_complete_v1(
    p_scope_schema name, p_export_id uuid, p_lease_owner uuid,
    p_attempt_generation bigint, p_object_version text,
    p_provider_request_id text, p_uploaded_at timestamptz,
    p_completed_at timestamptz
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_export analytics_dashboard_exports%ROWTYPE;
    v_next text;
BEGIN
    IF pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       p_lease_owner IS NULL OR
       p_lease_owner='00000000-0000-0000-0000-000000000000'::uuid OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.octet_length(p_object_version) NOT BETWEEN 1 AND 512 OR
       pg_catalog.octet_length(p_provider_request_id) NOT BETWEEN 1 AND 512 OR
       p_uploaded_at IS NULL OR NOT pg_catalog.isfinite(p_uploaded_at) OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       p_completed_at<p_uploaded_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export completion input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT item.* INTO v_export
    FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.state='running'
      AND item.lease_owner=p_lease_owner
      AND item.attempt_generation=p_attempt_generation
      AND item.effect_started_at IS NOT NULL
      AND item.object_version IS NULL
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics dashboard export completion changed';
    END IF;
    v_next:=CASE
        WHEN v_export.cancellation_requested_at IS NOT NULL THEN 'deleting'
        WHEN v_export.expires_at<=p_completed_at THEN 'expired'
        ELSE 'ready'
    END;
    UPDATE analytics_dashboard_export_attempt_receipts
    SET outcome='ready',object_version=p_object_version,
        provider_request_id=p_provider_request_id,completed_at=p_completed_at
    WHERE export_id=p_export_id
      AND attempt_generation=p_attempt_generation
      AND outcome='started';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics dashboard export receipt changed';
    END IF;
    UPDATE analytics_dashboard_exports
    SET state=v_next,object_version=p_object_version,
        provider_request_id=p_provider_request_id,
        ready_at=CASE WHEN v_next='ready' THEN p_completed_at ELSE NULL END,
        failure_code=NULL,
        row_revision=row_revision+1,lease_owner=NULL,lease_until=NULL,
        updated_at=p_completed_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    RETURN v_next;
END;
$function$;

CREATE FUNCTION analytics_dashboard_export_restart_missing_v1(
    p_scope_schema name, p_export_id uuid, p_lease_owner uuid,
    p_attempt_generation bigint, p_observed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_attempt_generation NOT BETWEEN 1 AND 3 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export restart input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    UPDATE analytics_dashboard_export_attempt_receipts
    SET outcome='missing',completed_at=p_observed_at
    WHERE export_id=p_export_id
      AND attempt_generation=p_attempt_generation
      AND outcome='started';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics dashboard export missing receipt changed';
    END IF;
    UPDATE analytics_dashboard_exports
    SET state=CASE WHEN attempt_generation>=3 THEN 'failed' ELSE 'queued' END,
        failure_code=CASE WHEN attempt_generation>=3
                          THEN 'MAX_ATTEMPTS' ELSE NULL END,
        row_revision=row_revision+1,lease_owner=NULL,lease_until=NULL,
        effect_started_at=NULL,object_key=NULL,object_version=NULL,
        key_id=NULL,wrapped_dek=NULL,wrap_nonce=NULL,header_sha256=NULL,
        ciphertext_sha256=NULL,plaintext_sha256=NULL,
        ciphertext_bytes=NULL,plaintext_bytes=NULL,row_count=NULL,
        filename=NULL,content_type=NULL,provider_request_id=NULL,
        updated_at=p_observed_at
    WHERE analytics_dashboard_exports.export_id=p_export_id
      AND state='running' AND lease_owner=p_lease_owner
      AND attempt_generation=p_attempt_generation
      AND effect_started_at IS NOT NULL;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics dashboard export restart changed';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_complete_v1(
    name,uuid,uuid,bigint,text,text,timestamptz,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_restart_missing_v1(
    name,uuid,uuid,bigint,timestamptz
) FROM PUBLIC;
