-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_guard_dashboard_export_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE v_transition boolean;
BEGIN
    IF TG_OP='DELETE' THEN
        RAISE EXCEPTION USING ERRCODE='P6104',
            MESSAGE='analytics dashboard export is immutable';
    END IF;
    IF NEW.export_id IS DISTINCT FROM OLD.export_id OR
       NEW.dashboard_id IS DISTINCT FROM OLD.dashboard_id OR
       NEW.dashboard_version_id IS DISTINCT FROM OLD.dashboard_version_id OR
       NEW.job_id IS DISTINCT FROM OLD.job_id OR
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
        WHEN 'expired' THEN NEW.state IN ('expired','deleted')
        WHEN 'deleted' THEN NEW.state='deleted'
        ELSE false
    END;
    IF NOT v_transition THEN
        RAISE EXCEPTION USING ERRCODE='P6104',
            MESSAGE='analytics dashboard export transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION analytics_guard_dashboard_export_attempt_receipt_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
    IF TG_OP='DELETE' OR OLD.outcome<>'started' OR
       NEW.outcome NOT IN ('missing','ready','unknown','failed','deleted') OR
       NEW.export_id IS DISTINCT FROM OLD.export_id OR
       NEW.attempt_generation IS DISTINCT FROM OLD.attempt_generation OR
       NEW.object_key IS DISTINCT FROM OLD.object_key OR
       NEW.key_id IS DISTINCT FROM OLD.key_id OR
       NEW.wrapped_dek IS DISTINCT FROM OLD.wrapped_dek OR
       NEW.wrap_nonce IS DISTINCT FROM OLD.wrap_nonce OR
       NEW.header_sha256 IS DISTINCT FROM OLD.header_sha256 OR
       NEW.ciphertext_sha256 IS DISTINCT FROM OLD.ciphertext_sha256 OR
       NEW.plaintext_sha256 IS DISTINCT FROM OLD.plaintext_sha256 OR
       NEW.ciphertext_bytes IS DISTINCT FROM OLD.ciphertext_bytes OR
       NEW.plaintext_bytes IS DISTINCT FROM OLD.plaintext_bytes OR
       NEW.row_count IS DISTINCT FROM OLD.row_count OR
       NEW.started_at IS DISTINCT FROM OLD.started_at OR
       NEW.completed_at IS NULL OR NEW.completed_at<OLD.started_at OR
       (NEW.outcome='ready' AND
        (NEW.object_version IS NULL OR NEW.provider_request_id IS NULL)) THEN
        RAISE EXCEPTION USING ERRCODE='P6104',
            MESSAGE='analytics dashboard export attempt receipt is immutable';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE TRIGGER analytics_dashboard_exports_guard
BEFORE UPDATE OR DELETE ON analytics_dashboard_exports
FOR EACH ROW EXECUTE FUNCTION analytics_guard_dashboard_export_v1();
ALTER TABLE analytics_dashboard_exports
    ENABLE ALWAYS TRIGGER analytics_dashboard_exports_guard;
CREATE TRIGGER analytics_dashboard_export_attempt_receipts_guard
BEFORE UPDATE OR DELETE ON analytics_dashboard_export_attempt_receipts
FOR EACH ROW EXECUTE FUNCTION analytics_guard_dashboard_export_attempt_receipt_v1();
ALTER TABLE analytics_dashboard_export_attempt_receipts
    ENABLE ALWAYS TRIGGER analytics_dashboard_export_attempt_receipts_guard;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_guard_dashboard_export_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_guard_dashboard_export_attempt_receipt_v1() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION analytics_dashboard_export_submit_v1(
    name,uuid,uuid,text,bytea,uuid,uuid,uuid,text,bytea,
    bigint,bigint,bigint,bigint,bigint,timestamptz,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_snapshot_v1(
    name,uuid,uuid,uuid,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_delete_v1(
    name,uuid,uuid,uuid,text,bytea,uuid,uuid,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_content_v1(
    name,uuid,uuid,uuid,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_claim_v1(
    name,uuid,uuid,uuid,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_effect_start_v1(
    name,uuid,uuid,bigint,text,text,bytea,bytea,bytea,bytea,bytea,
    bigint,bigint,bigint,text,text,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_complete_v1(
    name,uuid,uuid,bigint,text,text,timestamptz,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_restart_missing_v1(
    name,uuid,uuid,bigint,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_outcome_unknown_v1(
    name,uuid,uuid,bigint,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_fail_render_v1(
    name,uuid,uuid,bigint,text,boolean,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_delete_claim_v1(
    name,uuid,uuid,uuid,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_delete_complete_v1(
    name,uuid,uuid,uuid,bigint,text,text,bytea,bytea,timestamptz,timestamptz
) TO forgecrud_runtime;
