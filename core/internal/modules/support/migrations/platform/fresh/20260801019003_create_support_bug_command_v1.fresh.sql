-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.support_create_bug_v1(
    p_id uuid, p_tenant_id uuid, p_tenant_slug text, p_tenant_name text,
    p_actor_id uuid, p_reporter_name text, p_reporter_email text,
    p_authorization_version bigint, p_authority_clock bigint,
    p_title text, p_description text, p_category text, p_severity text,
    p_page_url text, p_user_agent text, p_browser_info jsonb,
    p_console_errors jsonb, p_network_info jsonb, p_steps_to_reproduce text,
    p_idempotency_key text, p_request_sha256 bytea,
    p_observed_at timestamp(6) with time zone
) RETURNS TABLE(response_document jsonb, replayed boolean)
LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_partition text;
    v_receipt platform.support_command_receipts%ROWTYPE;
    v_report platform.support_bug_reports%ROWTYPE;
    v_document jsonb;
BEGIN
    IF p_id IS NULL OR pg_catalog.uuid_extract_version(p_id) IS DISTINCT FROM 7 OR
       p_tenant_id IS NULL OR p_actor_id IS NULL OR
       p_authorization_version < 1 OR p_authority_clock < 1 OR
       p_tenant_slug IS NULL OR p_tenant_slug <> pg_catalog.btrim(p_tenant_slug) OR pg_catalog.octet_length(p_tenant_slug) NOT BETWEEN 1 AND 100 OR
       p_tenant_name IS NULL OR p_tenant_name <> pg_catalog.btrim(p_tenant_name) OR pg_catalog.octet_length(p_tenant_name) NOT BETWEEN 1 AND 200 OR
       p_reporter_name IS NULL OR pg_catalog.octet_length(p_reporter_name) > 200 OR
       p_reporter_email IS NULL OR p_reporter_email <> pg_catalog.lower(p_reporter_email) OR pg_catalog.octet_length(p_reporter_email) > 255 OR
       p_title IS NULL OR p_title <> pg_catalog.btrim(p_title) OR pg_catalog.octet_length(p_title) NOT BETWEEN 3 AND 500 OR
       p_description IS NULL OR p_description <> pg_catalog.btrim(p_description) OR pg_catalog.octet_length(p_description) NOT BETWEEN 10 AND 8000 OR
       p_category IS NULL OR p_category NOT IN ('bug','feature_request','performance','ui_issue','data_issue','other') OR
       p_severity IS NULL OR p_severity NOT IN ('critical','high','medium','low') OR
       p_page_url IS NULL OR pg_catalog.octet_length(p_page_url) > 2048 OR
       p_user_agent IS NULL OR pg_catalog.octet_length(p_user_agent) > 1024 OR
       p_browser_info IS NULL OR pg_catalog.jsonb_typeof(p_browser_info) <> 'object' OR pg_catalog.octet_length(p_browser_info::text) > 16384 OR
       p_console_errors IS NULL OR pg_catalog.jsonb_typeof(p_console_errors) <> 'array' OR pg_catalog.jsonb_array_length(p_console_errors) > 50 OR pg_catalog.octet_length(p_console_errors::text) > 32768 OR
       p_network_info IS NULL OR pg_catalog.jsonb_typeof(p_network_info) <> 'array' OR pg_catalog.jsonb_array_length(p_network_info) > 50 OR pg_catalog.octet_length(p_network_info::text) > 32768 OR
       p_steps_to_reproduce IS NULL OR pg_catalog.octet_length(p_steps_to_reproduce) > 8000 OR
       p_idempotency_key IS NULL OR pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 16 AND 200 OR
       p_request_sha256 IS NULL OR pg_catalog.octet_length(p_request_sha256) <> 32 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) OR
       pg_catalog.date_trunc('microseconds', p_observed_at) <> p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6501', MESSAGE = 'support bug command invalid';
    END IF;
    v_partition := 'actor:' || p_actor_id::text;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'support.bug.create:' || v_partition || ':' || p_idempotency_key, 0));
    SELECT receipt.* INTO v_receipt FROM platform.support_command_receipts AS receipt
    WHERE receipt.operation = 'support.bug.create'
      AND receipt.principal_partition = v_partition
      AND receipt.idempotency_key = p_idempotency_key;
    IF FOUND THEN
        IF v_receipt.request_sha256 <> p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6502', MESSAGE = 'support idempotency mismatch';
        END IF;
        RETURN QUERY SELECT v_receipt.response_document, true;
        RETURN;
    END IF;
    INSERT INTO platform.support_bug_reports(
        id, tenant_id, tenant_slug, tenant_name, reporter_id, reporter_name,
        reporter_email, title, description, category, severity, priority,
        page_url, user_agent, browser_info, console_errors, network_info,
        steps_to_reproduce, status, revision, created_at, updated_at)
    VALUES (p_id, p_tenant_id, p_tenant_slug, p_tenant_name, p_actor_id,
        p_reporter_name, p_reporter_email, p_title, p_description, p_category,
        p_severity, 'normal', p_page_url, p_user_agent, p_browser_info,
        p_console_errors, p_network_info, p_steps_to_reproduce, 'open', 1,
        p_observed_at, p_observed_at)
    RETURNING * INTO v_report;
    v_document := pg_catalog.to_jsonb(v_report);
    INSERT INTO platform.support_command_receipts(
        receipt_id, operation, principal_partition, idempotency_key,
        request_sha256, result_id, result_revision, response_document, created_at)
    VALUES (pg_catalog.uuidv7(), 'support.bug.create', v_partition,
        p_idempotency_key, p_request_sha256, p_id, 1, v_document, p_observed_at);
    RETURN QUERY SELECT v_document, false;
END;
$function$;

REVOKE ALL ON FUNCTION platform.support_create_bug_v1(uuid,uuid,text,text,uuid,text,text,bigint,bigint,text,text,text,text,text,text,jsonb,jsonb,jsonb,text,text,bytea,timestamp with time zone) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.support_create_bug_v1(uuid,uuid,text,text,uuid,text,text,bigint,bigint,text,text,text,text,text,text,jsonb,jsonb,jsonb,text,text,bytea,timestamp with time zone) TO harpy_runtime;
-- +goose StatementEnd
