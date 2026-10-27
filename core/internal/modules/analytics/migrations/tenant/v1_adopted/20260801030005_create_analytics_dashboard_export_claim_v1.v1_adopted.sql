-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_worker_snapshot_v1(
    p_export_id uuid, p_disposition text
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
BEGIN
    RETURN QUERY
    SELECT p_disposition,item.state,item.actor_id,item.export_id,
           item.dashboard_id,item.dashboard_version_id,item.job_id,
           item.token_version,item.authorization_version,
           item.authorization_clock,item.dashboard_manifest_sha256,
           item.pin_sha256,item.format,item.parameters_document,
           item.pinned_widgets,item.maximum_rows,
           item.maximum_plaintext_bytes,item.expires_at,
           item.attempt_generation,COALESCE(item.lease_owner,
               '00000000-0000-0000-0000-000000000000'::uuid),
           item.effect_started_at IS NOT NULL,COALESCE(item.object_key,''),
           COALESCE(item.key_id,''),COALESCE(item.wrapped_dek,''::bytea),
           COALESCE(item.wrap_nonce,''::bytea),
           COALESCE(item.header_sha256,''::bytea),
           COALESCE(item.ciphertext_sha256,''::bytea),
           COALESCE(item.plaintext_sha256,''::bytea),
           COALESCE(item.ciphertext_bytes,0),
           COALESCE(item.plaintext_bytes,0),COALESCE(item.row_count,0),
           COALESCE(item.content_type,''),COALESCE(item.filename,'')
    FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id;
END;
$function$;

CREATE FUNCTION analytics_dashboard_export_claim_v1(
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
    v_disposition text;
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
    SELECT item.* INTO v_export
    FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.job_id=p_job_id
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
        UPDATE analytics_dashboard_exports
        SET state=CASE
                WHEN analytics_dashboard_exports.expires_at<=p_observed_at THEN 'expired'
                WHEN analytics_dashboard_exports.effect_started_at IS NOT NULL THEN 'deleting'
                ELSE 'cancelled'
            END,
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
        UPDATE analytics_dashboard_exports
        SET lease_owner=p_lease_owner,
            lease_until=p_observed_at+INTERVAL '60 seconds',
            row_revision=row_revision+1,updated_at=p_observed_at
        WHERE analytics_dashboard_exports.export_id=p_export_id;
        v_disposition:='reconcile';
    ELSIF v_export.attempt_generation>=3 THEN
        UPDATE analytics_dashboard_exports
        SET state='failed',failure_code='MAX_ATTEMPTS',
            row_revision=row_revision+1,lease_owner=NULL,lease_until=NULL,
            updated_at=p_observed_at
        WHERE analytics_dashboard_exports.export_id=p_export_id;
        v_disposition:='terminal';
    ELSE
        UPDATE analytics_dashboard_exports
        SET state='running',attempt_generation=
                analytics_dashboard_exports.attempt_generation+1,
            lease_owner=p_lease_owner,
            lease_until=p_observed_at+INTERVAL '60 seconds',
            failure_code=NULL,row_revision=row_revision+1,updated_at=p_observed_at
        WHERE analytics_dashboard_exports.export_id=p_export_id;
        v_disposition:='execute';
    END IF;
    RETURN QUERY SELECT * FROM
        analytics_dashboard_export_worker_snapshot_v1(p_export_id,v_disposition);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_worker_snapshot_v1(
    uuid,text
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_claim_v1(
    name,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
