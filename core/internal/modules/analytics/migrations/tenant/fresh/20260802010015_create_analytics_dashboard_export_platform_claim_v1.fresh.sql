-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_platform_claim_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_generation bigint,p_lease_owner uuid
)
RETURNS TABLE(
    claim_state text,attempt_id uuid,claimed_at timestamptz,
    lease_until timestamptz,disposition text,state text,actor_id uuid,
    export_id uuid,dashboard_id uuid,dashboard_version_id uuid,job_id uuid,
    token_version bigint,authorization_version bigint,
    authorization_clock bigint,dashboard_manifest_sha256 bytea,
    pin_sha256 bytea,format text,parameters_document bytea,
    pinned_widgets jsonb,maximum_rows bigint,maximum_plaintext_bytes bigint,
    expires_at timestamptz,attempt_generation bigint,lease_owner uuid,
    effect_started boolean,object_key text,key_id text,wrapped_dek bytea,
    wrap_nonce bytea,header_sha256 bytea,ciphertext_sha256 bytea,
    plaintext_sha256 bytea,ciphertext_bytes bigint,plaintext_bytes bigint,
    row_count bigint,content_type text,filename text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_export analytics_dashboard_exports%ROWTYPE;
    v_claim_state text;
    v_attempt_id uuid;
    v_claimed_at timestamptz;
    v_lease_until timestamptz;
    v_replay boolean;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row
    WHERE identity_row.singleton
      AND identity_row.schema_name=pg_catalog.current_schema();
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       p_attempt_generation IS NULL OR p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.uuid_extract_version(p_lease_owner) IS DISTINCT FROM 7 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export claim scope is invalid';
    END IF;
    SELECT platform_claim.claim_state,platform_claim.attempt_id,
           platform_claim.claimed_at,platform_claim.lease_until
    INTO v_claim_state,v_attempt_id,v_claimed_at,v_lease_until
    FROM platform.platform_async_analytics_dashboard_export_claim_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_generation,p_lease_owner
    ) AS platform_claim;
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id
      AND item.job_id=p_job_id FOR UPDATE;
    IF NOT FOUND OR v_export.expires_at<=pg_catalog.clock_timestamp() OR
       v_export.cancellation_requested_at IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='Analytics dashboard export candidate changed';
    END IF;
    v_replay:=v_export.state IS NOT DISTINCT FROM 'running' AND
        v_export.quota_state IS NOT DISTINCT FROM 'held' AND
        v_export.attempt_generation IS NOT DISTINCT FROM p_attempt_generation AND
        v_export.lease_owner IS NOT DISTINCT FROM p_lease_owner AND
        v_export.platform_attempt_id IS NOT NULL;
    IF NOT v_replay AND (
       v_export.state IS DISTINCT FROM 'queued' OR
       v_export.quota_state IS DISTINCT FROM 'none' OR
       v_export.attempt_generation+1 IS DISTINCT FROM p_attempt_generation OR
       v_export.platform_attempt_id IS NOT NULL OR
       v_export.platform_effect_id IS NOT NULL OR
       v_export.effect_started_at IS NOT NULL) THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='Analytics dashboard export candidate changed';
    END IF;
    IF v_claim_state IS NOT DISTINCT FROM 'quota_wait' THEN
        IF v_replay OR v_attempt_id IS NOT NULL OR v_claimed_at IS NOT NULL OR
           v_lease_until IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export claim replay lost quota';
        END IF;
        claim_state:='quota_wait'; RETURN NEXT; RETURN;
    ELSIF v_claim_state IS DISTINCT FROM 'claimed' OR
          pg_catalog.uuid_extract_version(v_attempt_id) IS DISTINCT FROM 7 OR
          v_claimed_at IS NULL OR NOT pg_catalog.isfinite(v_claimed_at) OR
          v_lease_until IS NULL OR NOT pg_catalog.isfinite(v_lease_until) OR
          v_lease_until<=v_claimed_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export platform claim is invalid';
    END IF;
    IF v_replay THEN
        IF v_export.platform_attempt_id IS DISTINCT FROM v_attempt_id OR
           v_export.lease_until IS DISTINCT FROM v_lease_until THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export claim replay differs';
        END IF;
        PERFORM 1 FROM analytics_dashboard_export_platform_attempts AS attempt
        WHERE attempt.export_id=p_export_id
          AND attempt.attempt_generation=p_attempt_generation
          AND attempt.platform_attempt_id=v_attempt_id
          AND attempt.job_id=p_job_id AND attempt.lease_owner=p_lease_owner
          AND attempt.outcome='running' FOR UPDATE;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export attempt evidence is missing';
        END IF;
    ELSE
        UPDATE analytics_dashboard_exports
        SET state='running',attempt_generation=p_attempt_generation,
            quota_state='held',lease_owner=p_lease_owner,
            lease_until=v_lease_until,platform_attempt_id=v_attempt_id,
            failure_code=NULL,row_revision=row_revision+1,
            updated_at=GREATEST(updated_at,v_claimed_at)
        WHERE analytics_dashboard_exports.export_id=p_export_id;
        INSERT INTO analytics_dashboard_export_platform_attempts(
            export_id,attempt_generation,job_id,platform_attempt_id,lease_owner,
            claimed_at,lease_until,outcome
        ) VALUES (
            p_export_id,p_attempt_generation,p_job_id,v_attempt_id,p_lease_owner,
            v_claimed_at,v_lease_until,'running'
        );
    END IF;
    RETURN QUERY
    SELECT 'claimed'::text,v_attempt_id,v_claimed_at,v_lease_until,snapshot.*
    FROM analytics_dashboard_export_worker_snapshot_v1(
        p_export_id,'execute'
    ) AS snapshot;
END;
$function$;

-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_platform_claim_v1(
    name,uuid,uuid,uuid,bigint,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_platform_claim_v1(
    name,uuid,uuid,uuid,bigint,uuid) TO forgecrud_runtime;
