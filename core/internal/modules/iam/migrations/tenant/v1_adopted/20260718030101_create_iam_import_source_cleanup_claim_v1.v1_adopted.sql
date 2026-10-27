-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_import_source_cleanup_claim_v1(
    p_scope_schema name,p_tenant_id uuid,p_lease_owner uuid,
    p_observed_at timestamptz
)
RETURNS TABLE(
    source_id uuid,actor_id uuid,resource_kind text,object_key text,
    object_version text,cleanup_attempt integer,lease_generation bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_source iam_system_import_sources%ROWTYPE;
    v_request iam_system_record_imports%ROWTYPE;
    v_job_state text;
    v_job_completed_at timestamptz;
    v_eligible boolean;
BEGIN
    SELECT identity.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity
    WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema();
    IF p_tenant_id IS NULL OR v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.uuid_extract_version(p_lease_owner) IS DISTINCT FROM 7 OR
       NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM import source cleanup claim is invalid';
    END IF;
    FOR v_source IN
        SELECT source.*
        FROM iam_system_import_sources AS source
        WHERE source.tenant_id=p_tenant_id
          AND source.state IN ('ready','claimed')
          AND ((source.cleanup_state='waiting' AND
                source.cleanup_next_at<=p_observed_at) OR
               (source.cleanup_state='running' AND
                source.cleanup_lease_until<=p_observed_at))
        ORDER BY source.cleanup_next_at,source.source_id
        FOR UPDATE SKIP LOCKED
        LIMIT 32
    LOOP
        v_eligible:=false;
        IF v_source.state='ready' THEN
            v_eligible:=v_source.delete_after<=p_observed_at;
        ELSE
            SELECT request.* INTO STRICT v_request
            FROM iam_system_record_imports AS request
            WHERE request.import_id=v_source.claimed_import_id
              AND request.source_id=v_source.source_id
              AND request.tenant_id=p_tenant_id
              AND request.actor_id=v_source.actor_id
              AND request.resource_kind=v_source.resource_kind;
            SELECT job.job_state,job.completed_at
            INTO v_job_state,v_job_completed_at
            FROM platform.platform_async_iam_import_cleanup_read_v1(
                p_tenant_id,v_request.import_id,v_request.job_id,
                v_request.resource_kind
            ) AS job;
            IF v_job_state NOT IN ('queued','running','succeeded','dead','cancelled') OR
               (v_job_state IN ('succeeded','dead','cancelled') AND
                v_job_completed_at IS NULL) OR
               (v_job_state='succeeded' AND v_request.state<>'completed') THEN
                RAISE EXCEPTION USING ERRCODE='P6047',
                    MESSAGE='IAM import source cleanup state is inconsistent';
            END IF;
            v_eligible:=v_job_state IN ('succeeded','dead','cancelled');
        END IF;
        IF NOT v_eligible THEN
            UPDATE iam_system_import_sources
            SET cleanup_state='waiting',
                cleanup_next_at=p_observed_at+interval '1 minute',
                cleanup_lease_owner=NULL,cleanup_lease_until=NULL
            WHERE iam_system_import_sources.source_id=v_source.source_id;
            CONTINUE;
        END IF;
        RETURN QUERY
        UPDATE iam_system_import_sources AS claimed
        SET cleanup_state='running',cleanup_attempts=claimed.cleanup_attempts+1,
            cleanup_lease_owner=p_lease_owner,
            cleanup_lease_generation=claimed.cleanup_lease_generation+1,
            cleanup_lease_until=p_observed_at+interval '60 seconds',
            cleanup_next_at=p_observed_at
        WHERE claimed.source_id=v_source.source_id
        RETURNING claimed.source_id,claimed.actor_id,claimed.resource_kind,
                  claimed.object_key,claimed.object_version,
                  claimed.cleanup_attempts,claimed.cleanup_lease_generation;
        RETURN;
    END LOOP;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_import_source_cleanup_claim_v1(
    name,uuid,uuid,timestamptz
) FROM PUBLIC;
