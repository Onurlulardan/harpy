-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_execute_retention_cycle_v1(
    p_scope_schema name,p_tenant_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_successor_job_id uuid,p_payload_reference text,p_payload_sha256 text,
    p_limit integer
)
RETURNS TABLE (
    successor_job_id uuid,successor_run_at timestamptz,
    successor_payload_reference text,successor_payload_sha256 text,
    successor_dedupe_key text,successor_priority integer,
    drafts_purged bigint,versions_purged bigint,releases_purged bigint,
    executions_purged bigint,audits_purged bigint,capabilities_purged bigint,
    legal_hold_blocked boolean,has_remaining boolean,checkpoint_revision bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_stored analytics_retention_receipts%ROWTYPE;
    v_now timestamptz;
    v_next_at timestamptz;
    v_next_key text;
    v_slice integer;
    v_draft_limit integer;
    v_query_count bigint;
    v_dashboard_count bigint;
    v_audit_limit integer;
    v_receipt_sha256 bytea;
    v_actual_tenant uuid;
BEGIN
    v_now := pg_catalog.clock_timestamp();
    IF pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_successor_job_id) IS DISTINCT FROM 7 OR
       p_tenant_id IS NULL OR
       p_tenant_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       p_payload_reference IS DISTINCT FROM 'analytics-retention/' || p_tenant_id::text OR
       p_payload_sha256 !~ '^[0-9a-f]{64}$' OR
       pg_catalog.decode(p_payload_sha256,'hex') IS DISTINCT FROM
           pg_catalog.sha256(pg_catalog.convert_to(p_payload_reference,'UTF8')) OR
       p_limit NOT BETWEEN 4 AND 100 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'analytics retention cycle identity is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_actual_tenant
    FROM tenancy_schema_identity AS identity_row
    WHERE identity_row.singleton
      AND identity_row.schema_name = pg_catalog.current_schema();
    IF v_actual_tenant IS DISTINCT FROM p_tenant_id THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics retention tenant differs';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:retention:' || p_tenant_id::text,6100));
    SELECT receipt.* INTO v_stored FROM analytics_retention_receipts AS receipt
    WHERE receipt.job_id = p_job_id FOR SHARE;
    IF FOUND THEN
        IF v_stored.tenant_id IS DISTINCT FROM p_tenant_id OR
           v_stored.payload_reference IS DISTINCT FROM p_payload_reference OR
           v_stored.payload_sha256 IS DISTINCT FROM pg_catalog.decode(p_payload_sha256,'hex') THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103',
                MESSAGE = 'analytics retention job identity was reused';
        END IF;
        RETURN QUERY SELECT v_stored.successor_job_id,v_stored.successor_run_at,
            v_stored.payload_reference,pg_catalog.encode(v_stored.payload_sha256,'hex'),
            v_stored.successor_dedupe_key,0,v_stored.drafts_purged,
            v_stored.versions_purged,v_stored.releases_purged,
            v_stored.executions_purged,v_stored.audits_purged,
            v_stored.capabilities_purged,v_stored.legal_hold_blocked,
            v_stored.has_remaining,v_stored.checkpoint_revision;
        RETURN;
    END IF;
    drafts_purged := 0; versions_purged := 0; releases_purged := 0;
    executions_purged := 0; audits_purged := 0; capabilities_purged := 0;
    legal_hold_blocked := EXISTS (
        SELECT 1 FROM analytics_legal_holds AS hold WHERE hold.state = 'active'
    );
    has_remaining := false;
    IF NOT legal_hold_blocked THEN
        INSERT INTO analytics_retention_contexts (
            backend_pid,transaction_id,job_id,opened_at
        ) VALUES (
            pg_catalog.pg_backend_pid(),pg_catalog.pg_current_xact_id(),p_job_id,v_now
        );
        v_slice := p_limit / 4;
        SELECT result.executions_purged,result.audits_purged,result.capabilities_purged
        INTO executions_purged,audits_purged,capabilities_purged
        FROM analytics_retention_purge_audits_v1(
            v_now - interval '400 days',v_now - interval '7 years',
            v_now - interval '30 days',v_slice
        ) AS result;
        v_audit_limit := executions_purged + audits_purged + capabilities_purged;
        has_remaining := v_audit_limit >= v_slice;
        releases_purged := analytics_retention_purge_releases_v1(
            v_now - interval '400 days',v_slice
        );
        has_remaining := has_remaining OR releases_purged >= v_slice;
        SELECT result.query_versions_purged,result.dashboard_versions_purged
        INTO v_query_count,v_dashboard_count
        FROM analytics_retention_purge_versions_v1(
            v_now - interval '400 days',v_slice
        ) AS result;
        versions_purged := v_query_count + v_dashboard_count;
        has_remaining := has_remaining OR versions_purged >= v_slice;
        v_draft_limit := p_limit - v_audit_limit::integer
            - releases_purged::integer - versions_purged::integer;
        SELECT result.query_drafts_purged,result.dashboard_drafts_purged
        INTO v_query_count,v_dashboard_count
        FROM analytics_retention_purge_drafts_v1(
            v_now - interval '30 days',v_draft_limit
        ) AS result;
        drafts_purged := v_query_count + v_dashboard_count;
        has_remaining := has_remaining OR drafts_purged >= v_draft_limit;
        DELETE FROM analytics_retention_contexts AS retention_context
        WHERE retention_context.backend_pid = pg_catalog.pg_backend_pid()
          AND retention_context.transaction_id = pg_catalog.pg_current_xact_id();
    END IF;
    UPDATE analytics_retention_checkpoints SET
        revision = revision + 1,cycle_count = cycle_count + 1,
        last_job_id = p_job_id,last_attempt_id = p_attempt_id,
        last_completed_at = v_now WHERE singleton
    RETURNING revision INTO checkpoint_revision;
    IF has_remaining AND NOT legal_hold_blocked THEN
        v_next_at := v_now + interval '1 minute';
    ELSE
        v_next_at := (pg_catalog.date_trunc('day',v_now AT TIME ZONE 'UTC')
            + interval '1 day 3 hours') AT TIME ZONE 'UTC';
    END IF;
    v_next_key := 'analytics:retention:' || p_successor_job_id::text;
    v_receipt_sha256 := pg_catalog.sha256(pg_catalog.convert_to(
        pg_catalog.jsonb_build_array(
            p_job_id,p_tenant_id,p_successor_job_id,v_next_at,drafts_purged,
            versions_purged,releases_purged,executions_purged,audits_purged,
            capabilities_purged,legal_hold_blocked,checkpoint_revision
        )::text,'UTF8'));
    INSERT INTO analytics_retention_receipts (
        job_id,attempt_id,tenant_id,payload_reference,payload_sha256,
        successor_job_id,successor_run_at,successor_dedupe_key,
        drafts_purged,versions_purged,releases_purged,executions_purged,
        audits_purged,capabilities_purged,legal_hold_blocked,has_remaining,
        checkpoint_revision,receipt_sha256,completed_at
    ) VALUES (
        p_job_id,p_attempt_id,p_tenant_id,p_payload_reference,
        pg_catalog.decode(p_payload_sha256,'hex'),p_successor_job_id,v_next_at,
        v_next_key,drafts_purged,versions_purged,releases_purged,
        executions_purged,audits_purged,capabilities_purged,
        legal_hold_blocked,has_remaining,checkpoint_revision,v_receipt_sha256,v_now
    );
    RETURN QUERY SELECT p_successor_job_id,v_next_at,p_payload_reference,
        p_payload_sha256,v_next_key,0,drafts_purged,versions_purged,
        releases_purged,executions_purged,audits_purged,capabilities_purged,
        legal_hold_blocked,has_remaining,checkpoint_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_execute_retention_cycle_v1(
    name,uuid,uuid,uuid,uuid,text,text,integer
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_execute_retention_cycle_v1(
    name,uuid,uuid,uuid,uuid,text,text,integer
) TO harpy_runtime;
