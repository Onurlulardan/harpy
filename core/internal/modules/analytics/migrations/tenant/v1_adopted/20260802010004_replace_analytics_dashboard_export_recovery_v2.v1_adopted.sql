-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION analytics_dashboard_export_restart_missing_v1(
    p_scope_schema name, p_export_id uuid, p_lease_owner uuid,
    p_attempt_generation bigint, p_observed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_export analytics_dashboard_exports%ROWTYPE;
BEGIN
    IF p_attempt_generation NOT BETWEEN 1 AND 3 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export restart input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT item.* INTO v_export
    FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.state='running'
      AND item.lease_owner=p_lease_owner
      AND item.attempt_generation=p_attempt_generation
      AND item.effect_started_at IS NOT NULL AND item.quota_state='held'
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics dashboard export restart changed';
    END IF;
    UPDATE analytics_dashboard_export_attempt_receipts
    SET outcome='missing',completed_at=p_observed_at
    WHERE export_id=p_export_id
      AND attempt_generation=p_attempt_generation AND outcome='started';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics dashboard export missing receipt changed';
    END IF;
    UPDATE platform.artifact_quota_counters
    SET reserved_ciphertext_bytes=reserved_ciphertext_bytes-2147483648,
        revision=revision+1,updated_at=p_observed_at
    WHERE tenant_id=v_export.tenant_id
      AND reserved_ciphertext_bytes>=2147483648;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6105',
            MESSAGE='analytics dashboard export restart quota changed';
    END IF;
    UPDATE analytics_dashboard_exports
    SET state=CASE WHEN attempt_generation>=3 THEN 'failed' ELSE 'queued' END,
        quota_state=CASE WHEN attempt_generation>=3
            THEN 'released' ELSE 'none' END,
        failure_code=CASE WHEN attempt_generation>=3
            THEN 'MAX_ATTEMPTS' ELSE NULL END,
        row_revision=row_revision+1,lease_owner=NULL,lease_until=NULL,
        effect_started_at=NULL,object_key=NULL,object_version=NULL,
        key_id=NULL,wrapped_dek=NULL,wrap_nonce=NULL,header_sha256=NULL,
        ciphertext_sha256=NULL,plaintext_sha256=NULL,
        ciphertext_bytes=NULL,plaintext_bytes=NULL,row_count=NULL,
        filename=NULL,content_type=NULL,provider_request_id=NULL,
        updated_at=p_observed_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    RETURN true;
END;
$function$;

CREATE OR REPLACE FUNCTION analytics_dashboard_export_outcome_unknown_v1(
    p_scope_schema name, p_export_id uuid, p_lease_owner uuid,
    p_attempt_generation bigint, p_observed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       p_lease_owner IS NULL OR
       p_lease_owner='00000000-0000-0000-0000-000000000000'::uuid OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export unknown outcome input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    UPDATE analytics_dashboard_exports
    SET failure_code='ARTIFACT_OUTCOME_UNKNOWN',lease_owner=NULL,
        lease_until=NULL,row_revision=row_revision+1,updated_at=p_observed_at
    WHERE export_id=p_export_id AND state='running'
      AND lease_owner=p_lease_owner
      AND attempt_generation=p_attempt_generation
      AND effect_started_at IS NOT NULL AND quota_state='held';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics dashboard export unknown outcome changed';
    END IF;
    RETURN true;
END;
$function$;

CREATE OR REPLACE FUNCTION analytics_dashboard_export_fail_render_v1(
    p_scope_schema name, p_export_id uuid, p_lease_owner uuid,
    p_attempt_generation bigint, p_failure_code text,
    p_retryable boolean, p_failed_at timestamptz
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
       pg_catalog.octet_length(p_failure_code) NOT BETWEEN 3 AND 64 OR
       p_retryable IS NULL OR p_failed_at IS NULL OR
       NOT pg_catalog.isfinite(p_failed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export failure input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT item.* INTO v_export
    FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.state='running'
      AND item.lease_owner=p_lease_owner
      AND item.attempt_generation=p_attempt_generation
      AND item.effect_started_at IS NULL AND item.quota_state='held'
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics dashboard export failure changed';
    END IF;
    v_next:=CASE WHEN p_retryable AND p_attempt_generation<3
                 THEN 'queued' ELSE 'failed' END;
    UPDATE platform.artifact_quota_counters
    SET reserved_ciphertext_bytes=reserved_ciphertext_bytes-2147483648,
        revision=revision+1,updated_at=p_failed_at
    WHERE tenant_id=v_export.tenant_id
      AND reserved_ciphertext_bytes>=2147483648;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6105',
            MESSAGE='analytics dashboard export failure quota changed';
    END IF;
    UPDATE analytics_dashboard_exports
    SET state=v_next,quota_state=CASE WHEN v_next='queued'
            THEN 'none' ELSE 'released' END,
        failure_code=CASE WHEN v_next='failed' THEN p_failure_code ELSE NULL END,
        lease_owner=NULL,lease_until=NULL,row_revision=row_revision+1,
        updated_at=p_failed_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    RETURN v_next;
END;
$function$;
-- +goose StatementEnd

