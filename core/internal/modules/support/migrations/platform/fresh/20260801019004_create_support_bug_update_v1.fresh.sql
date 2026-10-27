-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.support_update_bug_v1(
    p_actor_id uuid, p_authorization_version bigint, p_authority_clock bigint,
    p_platform_scope boolean, p_tenant_id uuid, p_id uuid,
    p_has_status boolean, p_status text, p_has_priority boolean, p_priority text,
    p_has_resolution boolean, p_resolution text,
    p_has_admin_notes boolean, p_admin_notes text,
    p_expected_revision bigint, p_idempotency_key text, p_request_sha256 bytea,
    p_observed_at timestamp(6) with time zone
) RETURNS TABLE(response_document jsonb, replayed boolean)
LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation text;
    v_partition text;
    v_receipt platform.support_command_receipts%ROWTYPE;
    v_report platform.support_bug_reports%ROWTYPE;
    v_document jsonb;
BEGIN
    IF p_actor_id IS NULL OR p_authorization_version < 1 OR p_authority_clock < 1 OR
       p_platform_scope IS NULL OR p_tenant_id IS NULL OR p_id IS NULL OR
       (NOT p_has_status AND NOT p_has_priority AND NOT p_has_resolution AND NOT p_has_admin_notes) OR
       (p_has_status AND (p_status IS NULL OR p_status NOT IN ('open','in_progress','resolved','closed','wont_fix','duplicate'))) OR
       (NOT p_has_status AND p_status IS NOT NULL) OR
       (p_has_priority AND (p_priority IS NULL OR p_priority NOT IN ('urgent','high','normal','low'))) OR
       (NOT p_has_priority AND p_priority IS NOT NULL) OR
       (p_has_resolution AND (p_resolution IS NULL OR pg_catalog.octet_length(p_resolution) > 8000)) OR
       (NOT p_has_resolution AND p_resolution IS NOT NULL) OR
       (p_has_admin_notes AND (p_admin_notes IS NULL OR pg_catalog.octet_length(p_admin_notes) > 8000)) OR
       (NOT p_has_admin_notes AND p_admin_notes IS NOT NULL) OR
       p_expected_revision < 1 OR
       p_idempotency_key IS NULL OR pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 16 AND 200 OR
       p_request_sha256 IS NULL OR pg_catalog.octet_length(p_request_sha256) <> 32 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) OR
       pg_catalog.date_trunc('microseconds', p_observed_at) <> p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6501', MESSAGE = 'support bug update invalid';
    END IF;
    v_operation := CASE WHEN p_platform_scope
        THEN 'support.bug.platform.update' ELSE 'support.bug.tenant.update' END;
    v_partition := 'actor:' || p_actor_id::text;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        v_operation || ':' || v_partition || ':' || p_idempotency_key, 0));
    SELECT receipt.* INTO v_receipt FROM platform.support_command_receipts AS receipt
    WHERE receipt.operation = v_operation
      AND receipt.principal_partition = v_partition
      AND receipt.idempotency_key = p_idempotency_key;
    IF FOUND THEN
        IF v_receipt.request_sha256 <> p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6502', MESSAGE = 'support idempotency mismatch';
        END IF;
        RETURN QUERY SELECT v_receipt.response_document, true;
        RETURN;
    END IF;
    SELECT report.* INTO v_report FROM platform.support_bug_reports AS report
    WHERE report.id = p_id AND report.tenant_id = p_tenant_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6504', MESSAGE = 'support bug missing';
    END IF;
    IF v_report.revision <> p_expected_revision THEN
        RAISE EXCEPTION USING ERRCODE = 'P6503', MESSAGE = 'support revision stale';
    END IF;
    UPDATE platform.support_bug_reports AS report
    SET status = CASE WHEN p_has_status THEN p_status ELSE report.status END,
        priority = CASE WHEN p_has_priority THEN p_priority ELSE report.priority END,
        resolution = CASE WHEN p_has_resolution THEN p_resolution ELSE report.resolution END,
        admin_notes = CASE WHEN p_has_admin_notes THEN p_admin_notes ELSE report.admin_notes END,
        resolved_at = CASE WHEN p_has_status AND p_status = 'resolved' THEN p_observed_at ELSE report.resolved_at END,
        closed_at = CASE WHEN p_has_status AND p_status = 'closed' THEN p_observed_at ELSE report.closed_at END,
        revision = report.revision + 1, updated_at = p_observed_at
    WHERE report.id = p_id AND report.tenant_id = p_tenant_id
    RETURNING * INTO v_report;
    v_document := pg_catalog.to_jsonb(v_report);
    INSERT INTO platform.support_command_receipts(
        receipt_id, operation, principal_partition, idempotency_key,
        request_sha256, result_id, result_revision, response_document, created_at)
    VALUES (pg_catalog.uuidv7(), v_operation, v_partition, p_idempotency_key,
        p_request_sha256, p_id, v_report.revision, v_document, p_observed_at);
    RETURN QUERY SELECT v_document, false;
END;
$function$;

REVOKE ALL ON FUNCTION platform.support_update_bug_v1(uuid,bigint,bigint,boolean,uuid,uuid,boolean,text,boolean,text,boolean,text,boolean,text,bigint,text,bytea,timestamp with time zone) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.support_update_bug_v1(uuid,bigint,bigint,boolean,uuid,uuid,boolean,text,boolean,text,boolean,text,boolean,text,bigint,text,bytea,timestamp with time zone) TO forgecrud_runtime;
-- +goose StatementEnd
