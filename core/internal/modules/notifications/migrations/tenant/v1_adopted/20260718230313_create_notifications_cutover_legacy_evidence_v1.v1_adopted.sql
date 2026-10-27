-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_v1_adoption_legacy_evidence_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_id uuid
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_operation_created timestamptz;
    v_zero_since timestamptz;
    v_state text;
    v_hit_count bigint;
    v_revision bigint;
    v_tenant_sha text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Notifications V1 legacy evidence principal denied';
    END IF;
    SELECT operation.created_at INTO STRICT v_operation_created
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id
      AND operation.scope_kind='tenant' AND operation.scope_id=p_scope_id
      AND operation.group_id='notifications' AND operation.domain='notifications';
    SELECT retirement.zero_hit_since,retirement.state,
        retirement.hit_count,retirement.revision
    INTO STRICT v_zero_since,v_state,v_hit_count,v_revision
    FROM platform.contract_retirements AS retirement
    WHERE retirement.contract_key='notifications_actor_stream_v1'
      AND retirement.owner_module='notifications' AND retirement.domain='notifications'
      AND retirement.remaining_scope_count=0;
    IF v_state NOT IN ('eligible','retired') OR v_zero_since IS NULL OR
       v_zero_since>v_operation_created OR EXISTS (
        SELECT 1 FROM notifications_compatibility_hits AS hit
        WHERE hit.contract_name='notifications_actor_stream_v1'
          AND hit.occurred_at>=v_zero_since
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Notifications V1 legacy connection evidence is not quiet';
    END IF;
    SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(COALESCE(
        pg_catalog.string_agg(pg_catalog.to_jsonb(hit)::text,',' ORDER BY
            hit.occurred_at,hit.id),''),'UTF8')),'hex') INTO v_tenant_sha
    FROM notifications_compatibility_hits AS hit
    WHERE hit.contract_name='notifications_actor_stream_v1';
    RETURN pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        'harpy/notifications/v1-legacy-evidence/v1:'||p_operation_id::text||':'||
        p_run_id::text||':'||p_scope_id::text||':'||v_state||':'||v_revision::text||':'||
        v_hit_count::text||':'||v_zero_since::text||':'||v_tenant_sha,
        'UTF8'
    )),'hex');
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_v1_adoption_legacy_evidence_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
