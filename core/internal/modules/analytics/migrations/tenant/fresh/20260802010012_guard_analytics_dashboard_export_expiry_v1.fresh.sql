-- +goose Up
ALTER TABLE analytics_dashboard_exports
    ADD CONSTRAINT analytics_dashboard_exports_expiry_job_separation_check CHECK (
        expiry_job_id IS NULL OR
        (expiry_job_id<>export_id AND expiry_job_id<>job_id)
    );

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION analytics_guard_dashboard_export_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_transition boolean;
    v_quota_transition boolean;
    v_expiry_assignment boolean;
BEGIN
    IF TG_OP='DELETE' THEN
        RAISE EXCEPTION USING ERRCODE='P6104',
            MESSAGE='analytics dashboard export is immutable';
    END IF;
    v_expiry_assignment:=OLD.expiry_job_id IS NULL AND
        pg_catalog.uuid_extract_version(NEW.expiry_job_id)=7 AND
        OLD.state='queued' AND NEW.state='queued';
    IF NEW.export_id IS DISTINCT FROM OLD.export_id OR
       NEW.tenant_id IS DISTINCT FROM OLD.tenant_id OR
       NEW.dashboard_id IS DISTINCT FROM OLD.dashboard_id OR
       NEW.dashboard_version_id IS DISTINCT FROM OLD.dashboard_version_id OR
       NEW.job_id IS DISTINCT FROM OLD.job_id OR
       (NEW.expiry_job_id IS DISTINCT FROM OLD.expiry_job_id AND
        NOT v_expiry_assignment) OR
       NEW.actor_id IS DISTINCT FROM OLD.actor_id OR
       NEW.format IS DISTINCT FROM OLD.format OR
       NEW.idempotency_key IS DISTINCT FROM OLD.idempotency_key OR
       NEW.request_sha256 IS DISTINCT FROM OLD.request_sha256 OR
       NEW.token_version IS DISTINCT FROM OLD.token_version OR
       NEW.authorization_version IS DISTINCT FROM OLD.authorization_version OR
       NEW.authorization_clock IS DISTINCT FROM OLD.authorization_clock OR
       NEW.dashboard_manifest_sha256 IS DISTINCT FROM OLD.dashboard_manifest_sha256 OR
       NEW.pin_sha256 IS DISTINCT FROM OLD.pin_sha256 OR
       NEW.parameters_document IS DISTINCT FROM OLD.parameters_document OR
       NEW.parameters_sha256 IS DISTINCT FROM OLD.parameters_sha256 OR
       NEW.pinned_widgets IS DISTINCT FROM OLD.pinned_widgets OR
       NEW.widget_count IS DISTINCT FROM OLD.widget_count OR
       NEW.maximum_rows IS DISTINCT FROM OLD.maximum_rows OR
       NEW.maximum_plaintext_bytes IS DISTINCT FROM OLD.maximum_plaintext_bytes OR
       NEW.created_at IS DISTINCT FROM OLD.created_at OR
       NEW.expires_at IS DISTINCT FROM OLD.expires_at OR
       NEW.row_revision IS DISTINCT FROM OLD.row_revision+1 OR
       NEW.updated_at<OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='P6104',
            MESSAGE='analytics dashboard export pin is immutable';
    END IF;
    v_transition:=CASE OLD.state
        WHEN 'queued' THEN NEW.state IN ('queued','running','cancelled','expired')
        WHEN 'running' THEN NEW.state IN (
            'running','queued','ready','failed','cancelled','deleting','expired')
        WHEN 'ready' THEN NEW.state IN ('ready','deleting')
        WHEN 'failed' THEN NEW.state IN ('failed','deleted')
        WHEN 'cancelled' THEN NEW.state IN ('cancelled','deleting','deleted')
        WHEN 'deleting' THEN NEW.state IN ('deleting','deleted')
        WHEN 'expired' THEN NEW.state IN ('expired','deleting','deleted')
        WHEN 'deleted' THEN NEW.state='deleted'
        ELSE false
    END;
    v_quota_transition:=CASE OLD.quota_state
        WHEN 'none' THEN NEW.quota_state IN ('none','held','released')
        WHEN 'held' THEN NEW.quota_state IN ('held','none','used','released')
        WHEN 'used' THEN NEW.quota_state IN ('used','released')
        WHEN 'released' THEN NEW.quota_state='released'
        ELSE false
    END;
    IF NOT v_transition OR NOT v_quota_transition OR
       (NEW.state='ready' AND NEW.quota_state<>'used') OR
       (NEW.state='deleted' AND NEW.quota_state<>'released') OR
       (NEW.state='queued' AND NEW.quota_state<>'none') THEN
        RAISE EXCEPTION USING ERRCODE='P6104',
            MESSAGE='analytics dashboard export transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE EXECUTE ON FUNCTION analytics_dashboard_export_submit_v1(
    name,uuid,uuid,text,bytea,uuid,uuid,uuid,text,bytea,
    bigint,bigint,bigint,bigint,bigint,timestamptz,timestamptz
) FROM forgecrud_runtime;

