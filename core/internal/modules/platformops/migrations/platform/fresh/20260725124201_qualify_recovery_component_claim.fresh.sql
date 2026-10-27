-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.platformops_claim_next_recovery_set_v2(
    p_attempt_id uuid,
    p_lease_token uuid
)
RETURNS TABLE(
    attempt_id uuid,lease_token uuid,attempt_generation bigint,
    recovery_state text,recovery_set_id uuid,operation_id uuid,
    environment_id uuid,sequence bigint,primary_binding_id uuid,
    tenant_set_sha256 text,owner_binding_id uuid,
    owner_binding_generation bigint,database_lineage_id uuid,
    database_system_id uuid,
    recovery_signing_generation bigint,recovery_trust_bundle_sha256 text,
    database_now timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_set_id uuid;
    v_generation bigint;
    v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_backup'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'forgecrud-v2:core-database-backup-executor'
       OR pg_catalog.substring(p_attempt_id::text,15,1)<>'7'
       OR pg_catalog.substring(p_lease_token::text,15,1)<>'7' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery execution claim denied';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended('platformops:recovery-executor',0)
    );
    IF EXISTS (
        SELECT 1
        FROM platform.platformops_recovery_execution_attempts AS existing
        WHERE existing.attempt_id=p_attempt_id
          AND (existing.lease_token<>p_lease_token
               OR existing.lifecycle_state<>'active'
               OR existing.lease_expires_at<=v_now)
    ) THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='recovery execution claim conflicts';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM platform.platformops_recovery_execution_attempts AS existing
        WHERE existing.attempt_id=p_attempt_id
          AND existing.lease_token=p_lease_token
          AND existing.lifecycle_state='active'
          AND existing.lease_expires_at>v_now
    ) THEN
        RETURN QUERY
        SELECT execution.attempt_id,execution.lease_token,execution.generation,
               recovery.state,recovery.recovery_set_id,recovery.operation_id,
               recovery.environment_id,recovery.sequence,recovery.primary_binding_id,
               recovery.tenant_set_sha256,recovery.owner_binding_id,
               recovery.owner_binding_generation,recovery.database_lineage_id,
               lineage.current_database_system_id,
               recovery.recovery_signing_generation,
               recovery.recovery_trust_bundle_sha256,v_now
        FROM platform.platformops_recovery_execution_attempts AS execution
        JOIN platform.platformops_recovery_sets AS recovery
          ON recovery.recovery_set_id=execution.recovery_set_id
        JOIN platform.platformops_database_lineages AS lineage
          ON lineage.environment_id=recovery.environment_id
        WHERE execution.attempt_id=p_attempt_id;
        RETURN;
    END IF;
    SELECT recovery.recovery_set_id INTO v_set_id
    FROM platform.platformops_recovery_sets AS recovery
    LEFT JOIN platform.platformops_recovery_execution_attempts AS active
      ON active.recovery_set_id=recovery.recovery_set_id
     AND active.lifecycle_state='active'
    WHERE recovery.state='requested'
       OR (recovery.state IN ('holding','snapshotting','copying','verifying')
           AND active.lease_expires_at<=v_now)
    ORDER BY (recovery.state='requested') DESC,
             recovery.sequence,recovery.recovery_set_id
    LIMIT 1 FOR UPDATE OF recovery SKIP LOCKED;
    IF NOT FOUND THEN RETURN; END IF;
    UPDATE platform.platformops_recovery_execution_attempts AS execution
    SET lifecycle_state='superseded',completed_at=v_now
    WHERE execution.recovery_set_id=v_set_id
      AND execution.lifecycle_state='active'
      AND execution.lease_expires_at<=v_now;
    SELECT COALESCE(pg_catalog.max(execution.generation),0)+1
    INTO v_generation
    FROM platform.platformops_recovery_execution_attempts AS execution
    WHERE execution.recovery_set_id=v_set_id;
    UPDATE platform.platformops_recovery_sets AS recovery
    SET state=CASE WHEN recovery.state='requested' THEN 'holding'
                   ELSE recovery.state END,
        updated_at=v_now
    WHERE recovery.recovery_set_id=v_set_id;
    INSERT INTO platform.platformops_recovery_execution_attempts(
        attempt_id,recovery_set_id,generation,lease_token,lifecycle_state,
        claimed_recovery_state,lease_expires_at,heartbeat_at,started_at
    )
    SELECT p_attempt_id,recovery.recovery_set_id,v_generation,p_lease_token,
           'active',recovery.state,v_now+interval '2 minutes',v_now,v_now
    FROM platform.platformops_recovery_sets AS recovery
    WHERE recovery.recovery_set_id=v_set_id;
    RETURN QUERY
    SELECT execution.attempt_id,execution.lease_token,execution.generation,
           recovery.state,recovery.recovery_set_id,recovery.operation_id,
           recovery.environment_id,recovery.sequence,recovery.primary_binding_id,
           recovery.tenant_set_sha256,recovery.owner_binding_id,
           recovery.owner_binding_generation,recovery.database_lineage_id,
           lineage.current_database_system_id,
           recovery.recovery_signing_generation,
           recovery.recovery_trust_bundle_sha256,v_now
    FROM platform.platformops_recovery_execution_attempts AS execution
    JOIN platform.platformops_recovery_sets AS recovery
      ON recovery.recovery_set_id=execution.recovery_set_id
    JOIN platform.platformops_database_lineages AS lineage
      ON lineage.environment_id=recovery.environment_id
    WHERE execution.attempt_id=p_attempt_id;
END;
$function$;

CREATE OR REPLACE FUNCTION platform.platformops_claim_next_recovery_set_v3(
    p_attempt_id uuid,p_lease_token uuid
) RETURNS TABLE(
    attempt_id uuid,lease_token uuid,attempt_generation bigint,
    recovery_state text,recovery_set_id uuid,operation_id uuid,
    environment_id uuid,sequence bigint,primary_binding_id uuid,
    tenant_set_sha256 text,owner_binding_id uuid,
    owner_binding_generation bigint,database_lineage_id uuid,
    database_system_id uuid,recovery_signing_generation bigint,
    recovery_trust_bundle_sha256 text,component_inventory jsonb,
    database_now timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
RETURN QUERY
WITH claimed AS MATERIALIZED (
    SELECT * FROM platform.platformops_claim_next_recovery_set_v2(
        p_attempt_id,p_lease_token
    )
)
SELECT claimed.attempt_id,claimed.lease_token,claimed.attempt_generation,
       claimed.recovery_state,claimed.recovery_set_id,claimed.operation_id,
       claimed.environment_id,claimed.sequence,claimed.primary_binding_id,
       claimed.tenant_set_sha256,claimed.owner_binding_id,
       claimed.owner_binding_generation,claimed.database_lineage_id,
       claimed.database_system_id,claimed.recovery_signing_generation,
       claimed.recovery_trust_bundle_sha256,
       COALESCE((
           SELECT pg_catalog.jsonb_object_agg(
               component.component_kind,
               pg_catalog.jsonb_build_object(
                   'object_key',component.object_key,
                   'object_version_id',component.object_version_id,
                   'ciphertext_sha256',component.ciphertext_sha256,
                   'ciphertext_bytes',component.ciphertext_bytes
               )
           )
           FROM platform.platformops_recovery_components AS component
           WHERE component.recovery_set_id=claimed.recovery_set_id
       ),'{}'::jsonb),
       claimed.database_now
FROM claimed;
END;
$function$;
-- +goose StatementEnd
