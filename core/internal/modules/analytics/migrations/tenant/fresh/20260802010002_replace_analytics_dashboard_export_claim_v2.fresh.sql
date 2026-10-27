-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION analytics_dashboard_export_claim_v1(
    p_scope_schema name, p_export_id uuid, p_job_id uuid,
    p_lease_owner uuid, p_observed_at timestamptz
)
RETURNS TABLE (
    disposition text, state text, actor_id uuid, export_id uuid,
    dashboard_id uuid, dashboard_version_id uuid, job_id uuid,
    token_version bigint, authorization_version bigint,
    authorization_clock bigint, dashboard_manifest_sha256 bytea,
    pin_sha256 bytea, format text, parameters_document bytea,
    pinned_widgets jsonb, maximum_rows bigint,
    maximum_plaintext_bytes bigint, expires_at timestamptz,
    attempt_generation bigint, lease_owner uuid,
    effect_started boolean, object_key text, key_id text,
    wrapped_dek bytea, wrap_nonce bytea, header_sha256 bytea,
    ciphertext_sha256 bytea, plaintext_sha256 bytea,
    ciphertext_bytes bigint, plaintext_bytes bigint, row_count bigint,
    content_type text, filename text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_export analytics_dashboard_exports%ROWTYPE;
    v_tenant_id uuid;
    v_disposition text;
    v_release boolean:=false;
BEGIN
    IF pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       p_lease_owner IS NULL OR
       p_lease_owner='00000000-0000-0000-0000-000000000000'::uuid OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export claim input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row
    WHERE identity_row.singleton
      AND identity_row.schema_name=pg_catalog.current_schema();
    SELECT item.* INTO v_export
    FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.job_id=p_job_id
      AND item.tenant_id=v_tenant_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics dashboard export job was not found';
    END IF;
    IF v_export.state NOT IN ('queued','running') THEN
        RETURN QUERY SELECT * FROM
            analytics_dashboard_export_worker_snapshot_v1(p_export_id,'terminal');
        RETURN;
    END IF;
    IF v_export.expires_at<=p_observed_at OR
       v_export.cancellation_requested_at IS NOT NULL THEN
        v_release:=v_export.effect_started_at IS NULL AND
            v_export.quota_state='held';
        IF v_release THEN
            UPDATE platform.artifact_quota_counters
            SET reserved_ciphertext_bytes=reserved_ciphertext_bytes-2147483648,
                revision=revision+1,updated_at=p_observed_at
            WHERE tenant_id=v_tenant_id
              AND reserved_ciphertext_bytes>=2147483648;
            IF NOT FOUND THEN
                RAISE EXCEPTION USING ERRCODE='P6105',
                    MESSAGE='analytics dashboard export quota release changed';
            END IF;
        END IF;
        UPDATE analytics_dashboard_exports
        SET state=CASE
                WHEN analytics_dashboard_exports.expires_at<=p_observed_at
                     AND effect_started_at IS NULL THEN 'expired'
                WHEN effect_started_at IS NOT NULL THEN 'deleting'
                ELSE 'cancelled'
            END,
            quota_state=CASE WHEN v_release THEN 'released' ELSE quota_state END,
            row_revision=row_revision+1,lease_owner=NULL,lease_until=NULL,
            updated_at=p_observed_at
        WHERE analytics_dashboard_exports.export_id=p_export_id;
        RETURN QUERY SELECT * FROM
            analytics_dashboard_export_worker_snapshot_v1(p_export_id,'terminal');
        RETURN;
    END IF;
    IF v_export.state='running' AND v_export.lease_until>p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE='55P03',
            MESSAGE='analytics dashboard export lease is active';
    END IF;
    IF v_export.state='running' AND v_export.effect_started_at IS NOT NULL THEN
        IF v_export.quota_state<>'held' THEN
            RAISE EXCEPTION USING ERRCODE='P6105',
                MESSAGE='analytics dashboard export reconcile quota is missing';
        END IF;
        v_disposition:='reconcile';
    ELSIF v_export.attempt_generation>=3 THEN
        IF v_export.quota_state='held' THEN
            UPDATE platform.artifact_quota_counters
            SET reserved_ciphertext_bytes=reserved_ciphertext_bytes-2147483648,
                revision=revision+1,updated_at=p_observed_at
            WHERE tenant_id=v_tenant_id
              AND reserved_ciphertext_bytes>=2147483648;
            IF NOT FOUND THEN
                RAISE EXCEPTION USING ERRCODE='P6105',
                    MESSAGE='analytics dashboard export terminal quota changed';
            END IF;
        END IF;
        UPDATE analytics_dashboard_exports
        SET state='failed',failure_code='MAX_ATTEMPTS',quota_state='released',
            row_revision=row_revision+1,lease_owner=NULL,lease_until=NULL,
            updated_at=p_observed_at
        WHERE analytics_dashboard_exports.export_id=p_export_id;
        RETURN QUERY SELECT * FROM
            analytics_dashboard_export_worker_snapshot_v1(p_export_id,'terminal');
        RETURN;
    ELSE
        v_disposition:='execute';
    END IF;
    IF v_export.quota_state='none' THEN
        UPDATE platform.artifact_quota_counters
        SET reserved_ciphertext_bytes=reserved_ciphertext_bytes+2147483648,
            revision=revision+1,updated_at=p_observed_at
        WHERE tenant_id=v_tenant_id AND NOT fenced
          AND used_ciphertext_bytes+reserved_ciphertext_bytes+2147483648
              <=21474836480;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6105',
                MESSAGE='analytics dashboard export tenant quota is unavailable';
        END IF;
    ELSIF v_export.quota_state<>'held' THEN
        RAISE EXCEPTION USING ERRCODE='P6105',
            MESSAGE='analytics dashboard export quota state is invalid';
    END IF;
    UPDATE analytics_dashboard_exports
    SET state='running',
        attempt_generation=CASE WHEN v_disposition='execute'
            THEN analytics_dashboard_exports.attempt_generation+1
            ELSE analytics_dashboard_exports.attempt_generation END,
        quota_state='held',lease_owner=p_lease_owner,
        lease_until=p_observed_at+INTERVAL '60 seconds',
        failure_code=NULL,row_revision=row_revision+1,updated_at=p_observed_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    RETURN QUERY SELECT * FROM
        analytics_dashboard_export_worker_snapshot_v1(p_export_id,v_disposition);
END;
$function$;
-- +goose StatementEnd

