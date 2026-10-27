-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_platform_unknown_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_attempt_generation bigint,p_lease_owner uuid,
    p_effect_id uuid,p_effect_request_sha256 bytea,p_observed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_export analytics_dashboard_exports%ROWTYPE;
    v_attempt analytics_dashboard_export_platform_attempts%ROWTYPE;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row WHERE identity_row.singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR p_lease_owner IS NULL OR
       pg_catalog.octet_length(p_effect_request_sha256) IS DISTINCT FROM 32 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export unknown scope is invalid';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_unknown_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_attempt_generation,
        p_lease_owner,p_effect_id,p_effect_request_sha256,p_observed_at);
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id
      AND item.job_id=p_job_id FOR UPDATE;
    SELECT attempt.* INTO v_attempt
    FROM analytics_dashboard_export_platform_attempts AS attempt
    WHERE attempt.export_id=p_export_id
      AND attempt.attempt_generation=p_attempt_generation
      AND attempt.platform_attempt_id=p_attempt_id
      AND attempt.job_id=p_job_id AND attempt.lease_owner=p_lease_owner
    FOR UPDATE;
    IF v_export.export_id IS NULL OR v_attempt.export_id IS NULL OR
       v_attempt.platform_effect_id IS DISTINCT FROM p_effect_id OR
       v_attempt.platform_effect_request_sha256 IS DISTINCT FROM
           p_effect_request_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export unknown attempt is stale';
    END IF;
    IF v_attempt.outcome='unknown' THEN
        IF v_export.state IS DISTINCT FROM 'running' OR
           v_export.quota_state IS DISTINCT FROM 'held' OR
           v_export.platform_attempt_id IS DISTINCT FROM p_attempt_id OR
           v_export.platform_effect_id IS DISTINCT FROM p_effect_id OR
           v_export.platform_effect_request_sha256 IS DISTINCT FROM
               p_effect_request_sha256 OR
           v_export.attempt_generation IS DISTINCT FROM p_attempt_generation OR
           v_export.lease_owner IS NOT NULL OR
           v_export.lease_until IS NOT NULL OR
           v_export.failure_code IS DISTINCT FROM 'ARTIFACT_OUTCOME_UNKNOWN' OR
           v_export.updated_at IS DISTINCT FROM p_observed_at OR
           v_attempt.failure_class IS NOT NULL OR
           v_attempt.failure_code IS DISTINCT FROM 'JOB_OUTCOME_UNKNOWN' OR
           v_attempt.failure_sha256 IS NOT NULL OR
           v_attempt.finished_at IS DISTINCT FROM p_observed_at THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export unknown replay differs';
        END IF;
        PERFORM 1 FROM analytics_dashboard_export_attempt_receipts AS receipt
        WHERE receipt.export_id=p_export_id
          AND receipt.attempt_generation=p_attempt_generation
          AND receipt.outcome='unknown'
          AND receipt.platform_attempt_id=p_attempt_id
          AND receipt.platform_effect_id=p_effect_id
          AND receipt.platform_effect_request_sha256=p_effect_request_sha256
          AND receipt.completed_at=p_observed_at FOR SHARE;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export unknown receipt differs';
        END IF;
        RETURN true;
    END IF;
    IF v_attempt.outcome IS DISTINCT FROM 'running' OR
       v_export.state IS DISTINCT FROM 'running' OR
       v_export.quota_state IS DISTINCT FROM 'held' OR
       v_export.platform_attempt_id IS DISTINCT FROM p_attempt_id OR
       v_export.platform_effect_id IS DISTINCT FROM p_effect_id OR
       v_export.platform_effect_request_sha256 IS DISTINCT FROM
           p_effect_request_sha256 OR
       v_export.attempt_generation IS DISTINCT FROM p_attempt_generation OR
       v_export.lease_owner IS DISTINCT FROM p_lease_owner OR
       v_export.effect_started_at IS NULL OR
       p_observed_at<=v_export.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export unknown lease is stale';
    END IF;
    UPDATE analytics_dashboard_export_attempt_receipts
    SET outcome='unknown',completed_at=p_observed_at
    WHERE export_id=p_export_id
      AND attempt_generation=p_attempt_generation AND outcome='started';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export unknown receipt changed';
    END IF;
    UPDATE analytics_dashboard_export_platform_attempts
    SET outcome='unknown',failure_code='JOB_OUTCOME_UNKNOWN',
        finished_at=p_observed_at
    WHERE export_id=p_export_id
      AND attempt_generation=p_attempt_generation AND outcome='running';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export unknown attempt changed';
    END IF;
    UPDATE analytics_dashboard_exports
    SET failure_code='ARTIFACT_OUTCOME_UNKNOWN',lease_owner=NULL,
        lease_until=NULL,row_revision=row_revision+1,updated_at=p_observed_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    RETURN true;
END;
$function$;

CREATE FUNCTION analytics_dashboard_export_platform_spool_active_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,
    p_attempt_generation bigint,p_lease_owner uuid
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_export analytics_dashboard_exports%ROWTYPE;
    v_attempt analytics_dashboard_export_platform_attempts%ROWTYPE;
    v_state text;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row WHERE identity_row.singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR p_lease_owner IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export spool scope is invalid';
    END IF;
    v_state:=platform.platform_async_analytics_dashboard_export_spool_state_v1(
        p_tenant_id,p_export_id,p_attempt_generation,p_lease_owner);
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id FOR SHARE;
    SELECT attempt.* INTO v_attempt
    FROM analytics_dashboard_export_platform_attempts AS attempt
    WHERE attempt.export_id=p_export_id
      AND attempt.attempt_generation=p_attempt_generation
      AND attempt.lease_owner=p_lease_owner FOR SHARE;
    IF v_export.export_id IS NULL OR v_attempt.export_id IS NULL THEN
        RETURN false;
    END IF;
    IF v_state IS NOT DISTINCT FROM 'running' THEN
        RETURN v_attempt.outcome IS NOT DISTINCT FROM 'running' AND
            v_export.state IS NOT DISTINCT FROM 'running' AND
            v_export.quota_state IS NOT DISTINCT FROM 'held' AND
            v_export.platform_attempt_id IS NOT DISTINCT FROM
                v_attempt.platform_attempt_id AND
            v_export.lease_owner IS NOT DISTINCT FROM p_lease_owner;
    ELSIF v_state IS NOT DISTINCT FROM 'outcome_unknown' THEN
        RETURN v_attempt.outcome IS NOT DISTINCT FROM 'unknown' AND
            v_attempt.resolution_receipt_id IS NULL AND
            v_export.state IS NOT DISTINCT FROM 'running' AND
            v_export.quota_state IS NOT DISTINCT FROM 'held' AND
            v_export.platform_attempt_id IS NOT DISTINCT FROM
                v_attempt.platform_attempt_id AND
            v_export.lease_owner IS NULL;
    ELSIF v_state IS NOT DISTINCT FROM 'inactive' THEN
        RETURN false;
    END IF;
    RAISE EXCEPTION USING ERRCODE='P6048',
        MESSAGE='Analytics dashboard export spool state is invalid';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_platform_unknown_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_platform_spool_active_v1(
    name,uuid,uuid,bigint,uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_platform_unknown_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,timestamptz) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_platform_spool_active_v1(
    name,uuid,uuid,bigint,uuid) TO harpy_runtime;

REVOKE EXECUTE ON FUNCTION analytics_dashboard_export_claim_v1(
    name,uuid,uuid,uuid,timestamptz) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION analytics_dashboard_export_effect_start_v1(
    name,uuid,uuid,bigint,text,text,bytea,bytea,bytea,bytea,bytea,
    bigint,bigint,bigint,text,text,timestamptz) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION analytics_dashboard_export_complete_v1(
    name,uuid,uuid,bigint,text,text,timestamptz,timestamptz) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION analytics_dashboard_export_restart_missing_v1(
    name,uuid,uuid,bigint,timestamptz) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION analytics_dashboard_export_outcome_unknown_v1(
    name,uuid,uuid,bigint,timestamptz) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION analytics_dashboard_export_fail_render_v1(
    name,uuid,uuid,bigint,text,boolean,timestamptz) FROM harpy_runtime;
