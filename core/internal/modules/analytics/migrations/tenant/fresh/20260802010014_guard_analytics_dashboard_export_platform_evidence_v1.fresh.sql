-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION analytics_guard_dashboard_export_attempt_receipt_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
    IF TG_OP='DELETE' OR OLD.outcome<>'started' OR
       NEW.outcome NOT IN ('missing','ready','unknown','failed','requeued',
                           'dead','cancelled','deleted') OR
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

CREATE FUNCTION analytics_guard_dashboard_export_platform_evidence_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_attempt_assignment boolean;
    v_attempt_clear boolean;
    v_effect_assignment boolean;
    v_effect_clear boolean;
BEGIN
    IF TG_OP='DELETE' THEN
        RAISE EXCEPTION USING ERRCODE='P6104',
            MESSAGE='Analytics dashboard export platform evidence is immutable';
    END IF;
    v_attempt_assignment:=OLD.platform_attempt_id IS NULL AND
        pg_catalog.uuid_extract_version(NEW.platform_attempt_id)=7 AND
        NEW.attempt_generation=OLD.attempt_generation+1 AND
        OLD.quota_state='none' AND NEW.quota_state='held' AND NEW.state='running';
    v_attempt_clear:=OLD.platform_attempt_id IS NOT NULL AND
        NEW.platform_attempt_id IS NULL AND
        NEW.attempt_generation=OLD.attempt_generation AND
        OLD.quota_state='held' AND NEW.quota_state<>'held';
    IF NEW.platform_attempt_id IS DISTINCT FROM OLD.platform_attempt_id AND
       NOT (v_attempt_assignment OR v_attempt_clear) THEN
        RAISE EXCEPTION USING ERRCODE='P6104',
            MESSAGE='Analytics dashboard export platform attempt evidence changed';
    END IF;
    v_effect_assignment:=OLD.platform_effect_id IS NULL AND
        OLD.platform_effect_request_sha256 IS NULL AND
        pg_catalog.uuid_extract_version(NEW.platform_effect_id)=7 AND
        pg_catalog.octet_length(NEW.platform_effect_request_sha256)=32 AND
        NEW.platform_attempt_id=OLD.platform_attempt_id AND
        OLD.effect_started_at IS NULL AND NEW.effect_started_at IS NOT NULL AND
        NEW.quota_state='held';
    v_effect_clear:=OLD.platform_effect_id IS NOT NULL AND
        OLD.platform_effect_request_sha256 IS NOT NULL AND
        NEW.platform_effect_id IS NULL AND
        NEW.platform_effect_request_sha256 IS NULL AND v_attempt_clear;
    IF (NEW.platform_effect_id IS DISTINCT FROM OLD.platform_effect_id OR
        NEW.platform_effect_request_sha256 IS DISTINCT FROM
            OLD.platform_effect_request_sha256) AND
       NOT (v_effect_assignment OR v_effect_clear) THEN
        RAISE EXCEPTION USING ERRCODE='P6104',
            MESSAGE='Analytics dashboard export platform effect evidence changed';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION analytics_guard_dashboard_export_receipt_platform_evidence_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
    IF TG_OP='DELETE' OR
       NEW.platform_attempt_id IS DISTINCT FROM OLD.platform_attempt_id OR
       NEW.platform_effect_id IS DISTINCT FROM OLD.platform_effect_id OR
       NEW.platform_effect_request_sha256 IS DISTINCT FROM
            OLD.platform_effect_request_sha256 OR
       (OLD.platform_effect_receipt_id IS NOT NULL AND
        NEW.platform_effect_receipt_id IS DISTINCT FROM
            OLD.platform_effect_receipt_id) OR
       (OLD.platform_effect_result_sha256 IS NOT NULL AND
        NEW.platform_effect_result_sha256 IS DISTINCT FROM
            OLD.platform_effect_result_sha256) OR
       (OLD.uploaded_at IS NOT NULL AND
        NEW.uploaded_at IS DISTINCT FROM OLD.uploaded_at) OR
       ((NEW.platform_effect_receipt_id IS DISTINCT FROM
            OLD.platform_effect_receipt_id OR
         NEW.platform_effect_result_sha256 IS DISTINCT FROM
            OLD.platform_effect_result_sha256) AND
        (OLD.outcome<>'started' OR NEW.outcome='started')) THEN
        RAISE EXCEPTION USING ERRCODE='P6104',
            MESSAGE='Analytics dashboard export receipt platform evidence changed';
    END IF;
    RETURN NEW;
END;
$function$;
CREATE FUNCTION analytics_guard_dashboard_export_platform_attempt_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
    IF TG_OP='DELETE' OR
       NEW.export_id IS DISTINCT FROM OLD.export_id OR
       NEW.attempt_generation IS DISTINCT FROM OLD.attempt_generation OR
       NEW.job_id IS DISTINCT FROM OLD.job_id OR
       NEW.platform_attempt_id IS DISTINCT FROM OLD.platform_attempt_id OR
       NEW.lease_owner IS DISTINCT FROM OLD.lease_owner OR
       NEW.claimed_at IS DISTINCT FROM OLD.claimed_at OR
       NEW.platform_effect_id IS DISTINCT FROM OLD.platform_effect_id AND
            OLD.platform_effect_id IS NOT NULL OR
       NEW.platform_effect_request_sha256 IS DISTINCT FROM
            OLD.platform_effect_request_sha256 AND
            OLD.platform_effect_request_sha256 IS NOT NULL OR
       NEW.effect_started_at IS DISTINCT FROM OLD.effect_started_at AND
            OLD.effect_started_at IS NOT NULL OR
       OLD.outcome<>'running' AND NEW.outcome IS DISTINCT FROM OLD.outcome OR
       OLD.finished_at IS NOT NULL AND
            NEW.finished_at IS DISTINCT FROM OLD.finished_at OR
       OLD.platform_effect_receipt_id IS NOT NULL AND
            NEW.platform_effect_receipt_id IS DISTINCT FROM
                OLD.platform_effect_receipt_id OR
       OLD.platform_effect_result_sha256 IS NOT NULL AND
            NEW.platform_effect_result_sha256 IS DISTINCT FROM
                OLD.platform_effect_result_sha256 OR
       OLD.uploaded_at IS NOT NULL AND
            NEW.uploaded_at IS DISTINCT FROM OLD.uploaded_at OR
       OLD.failure_class IS NOT NULL AND
            NEW.failure_class IS DISTINCT FROM OLD.failure_class OR
       OLD.failure_code IS NOT NULL AND
            NEW.failure_code IS DISTINCT FROM OLD.failure_code OR
       OLD.failure_sha256 IS NOT NULL AND
            NEW.failure_sha256 IS DISTINCT FROM OLD.failure_sha256 OR
       OLD.resolution_receipt_id IS NOT NULL AND
            NEW.resolution_receipt_id IS DISTINCT FROM OLD.resolution_receipt_id OR
       OLD.resolution_kind IS NOT NULL AND
            NEW.resolution_kind IS DISTINCT FROM OLD.resolution_kind OR
       OLD.resolution_sha256 IS NOT NULL AND
            NEW.resolution_sha256 IS DISTINCT FROM OLD.resolution_sha256 OR
       OLD.resolved_at IS NOT NULL AND
            NEW.resolved_at IS DISTINCT FROM OLD.resolved_at THEN
        RAISE EXCEPTION USING ERRCODE='P6104',
            MESSAGE='Analytics dashboard export platform attempt is immutable';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE TRIGGER analytics_dashboard_export_platform_attempts_guard
BEFORE UPDATE OR DELETE ON analytics_dashboard_export_platform_attempts
FOR EACH ROW EXECUTE FUNCTION
    analytics_guard_dashboard_export_platform_attempt_v1();
ALTER TABLE analytics_dashboard_export_platform_attempts
    ENABLE ALWAYS TRIGGER analytics_dashboard_export_platform_attempts_guard;

REVOKE ALL ON FUNCTION
    analytics_guard_dashboard_export_platform_attempt_v1()
FROM PUBLIC;

CREATE TRIGGER analytics_dashboard_exports_platform_evidence_guard
BEFORE UPDATE OR DELETE ON analytics_dashboard_exports
FOR EACH ROW EXECUTE FUNCTION
    analytics_guard_dashboard_export_platform_evidence_v1();
ALTER TABLE analytics_dashboard_exports
    ENABLE ALWAYS TRIGGER analytics_dashboard_exports_platform_evidence_guard;

CREATE TRIGGER analytics_dashboard_export_receipts_platform_evidence_guard
BEFORE UPDATE OR DELETE ON analytics_dashboard_export_attempt_receipts
FOR EACH ROW EXECUTE FUNCTION
    analytics_guard_dashboard_export_receipt_platform_evidence_v1();
ALTER TABLE analytics_dashboard_export_attempt_receipts
    ENABLE ALWAYS TRIGGER
        analytics_dashboard_export_receipts_platform_evidence_guard;

REVOKE ALL ON FUNCTION
    analytics_guard_dashboard_export_platform_evidence_v1()
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    analytics_guard_dashboard_export_receipt_platform_evidence_v1()
FROM PUBLIC;
-- +goose StatementEnd
