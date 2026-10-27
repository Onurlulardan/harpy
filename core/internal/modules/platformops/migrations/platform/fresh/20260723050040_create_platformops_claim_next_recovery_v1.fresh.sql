-- +goose Up
CREATE TABLE platform.platformops_recovery_execution_attempts (
    attempt_id uuid PRIMARY KEY
        CHECK (pg_catalog.substring(attempt_id::text,15,1)='7'),
    recovery_set_id uuid NOT NULL
        REFERENCES platform.platformops_recovery_sets(recovery_set_id),
    generation bigint NOT NULL CHECK (generation > 0),
    lease_token uuid NOT NULL UNIQUE
        CHECK (pg_catalog.substring(lease_token::text,15,1)='7'),
    lifecycle_state text NOT NULL CHECK (
        lifecycle_state IN ('active','completed','failed','superseded')
    ),
    claimed_recovery_state text NOT NULL CHECK (
        claimed_recovery_state IN ('holding','snapshotting','copying','verifying')
    ),
    documents_hold_sha256 text CHECK (
        documents_hold_sha256 IS NULL
        OR documents_hold_sha256 ~ '^[0-9a-f]{64}$'
    ),
    documents_inventory_count bigint CHECK (
        documents_inventory_count IS NULL OR documents_inventory_count >= 0
    ),
    lease_expires_at timestamptz NOT NULL,
    heartbeat_at timestamptz NOT NULL,
    failure_code text CHECK (
        failure_code IS NULL OR failure_code ~ '^[a-z0-9_]{1,64}$'
    ),
    started_at timestamptz NOT NULL,
    completed_at timestamptz,
    UNIQUE (recovery_set_id,generation),
    CHECK ((documents_hold_sha256 IS NULL) =
           (documents_inventory_count IS NULL)),
    CHECK ((lifecycle_state IN ('completed','failed','superseded')) =
           (completed_at IS NOT NULL)),
    CHECK ((lifecycle_state='failed') = (failure_code IS NOT NULL))
);
CREATE UNIQUE INDEX platformops_recovery_execution_active_key
ON platform.platformops_recovery_execution_attempts(recovery_set_id)
WHERE lifecycle_state='active';

-- +goose StatementBegin
CREATE FUNCTION platform.platformops_claim_next_recovery_set_v2(
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
        SELECT 1 FROM platform.platformops_recovery_execution_attempts
        WHERE attempt_id=p_attempt_id
          AND (lease_token<>p_lease_token OR lifecycle_state<>'active'
               OR lease_expires_at<=v_now)
    ) THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='recovery execution claim conflicts';
    END IF;
    IF EXISTS (
        SELECT 1 FROM platform.platformops_recovery_execution_attempts
        WHERE attempt_id=p_attempt_id AND lease_token=p_lease_token
          AND lifecycle_state='active' AND lease_expires_at>v_now
    ) THEN
        RETURN QUERY
        SELECT attempt.attempt_id,attempt.lease_token,attempt.generation,
               recovery.state,recovery.recovery_set_id,recovery.operation_id,
               recovery.environment_id,recovery.sequence,recovery.primary_binding_id,
               recovery.tenant_set_sha256,recovery.owner_binding_id,
               recovery.owner_binding_generation,recovery.database_lineage_id,
               lineage.current_database_system_id,
               recovery.recovery_signing_generation,
               recovery.recovery_trust_bundle_sha256,v_now
        FROM platform.platformops_recovery_execution_attempts AS attempt
        JOIN platform.platformops_recovery_sets AS recovery
          ON recovery.recovery_set_id=attempt.recovery_set_id
        JOIN platform.platformops_database_lineages AS lineage
          ON lineage.environment_id=recovery.environment_id
        WHERE attempt.attempt_id=p_attempt_id;
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
    ORDER BY (recovery.state='requested') DESC,recovery.sequence,recovery.recovery_set_id
    LIMIT 1 FOR UPDATE OF recovery SKIP LOCKED;
    IF NOT FOUND THEN RETURN; END IF;
    UPDATE platform.platformops_recovery_execution_attempts
    SET lifecycle_state='superseded',completed_at=v_now
    WHERE recovery_set_id=v_set_id AND lifecycle_state='active'
      AND lease_expires_at<=v_now;
    SELECT COALESCE(pg_catalog.max(generation),0)+1 INTO v_generation
    FROM platform.platformops_recovery_execution_attempts
    WHERE recovery_set_id=v_set_id;
    UPDATE platform.platformops_recovery_sets
    SET state=CASE WHEN state='requested' THEN 'holding' ELSE state END,
        updated_at=v_now WHERE recovery_set_id=v_set_id;
    INSERT INTO platform.platformops_recovery_execution_attempts(
        attempt_id,recovery_set_id,generation,lease_token,lifecycle_state,
        claimed_recovery_state,lease_expires_at,heartbeat_at,started_at
    )
    SELECT p_attempt_id,recovery_set_id,v_generation,p_lease_token,'active',
           state,v_now+interval '2 minutes',v_now,v_now
    FROM platform.platformops_recovery_sets WHERE recovery_set_id=v_set_id;
    RETURN QUERY
    SELECT attempt.attempt_id,attempt.lease_token,attempt.generation,
           recovery.state,recovery.recovery_set_id,recovery.operation_id,
           recovery.environment_id,recovery.sequence,recovery.primary_binding_id,
           recovery.tenant_set_sha256,recovery.owner_binding_id,
           recovery.owner_binding_generation,recovery.database_lineage_id,
           lineage.current_database_system_id,
           recovery.recovery_signing_generation,recovery.recovery_trust_bundle_sha256,
           v_now
    FROM platform.platformops_recovery_execution_attempts AS attempt
    JOIN platform.platformops_recovery_sets AS recovery
      ON recovery.recovery_set_id=attempt.recovery_set_id
    JOIN platform.platformops_database_lineages AS lineage
      ON lineage.environment_id=recovery.environment_id
    WHERE attempt.attempt_id=p_attempt_id;
END;
$function$;

CREATE FUNCTION platform.platformops_heartbeat_recovery_execution_v1(
    p_recovery_set_id uuid,p_attempt_id uuid,p_lease_token uuid
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_backup'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'forgecrud-v2:core-database-backup-executor' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery execution heartbeat denied';
    END IF;
    UPDATE platform.platformops_recovery_execution_attempts
    SET heartbeat_at=v_now,lease_expires_at=v_now+interval '2 minutes'
    WHERE attempt_id=p_attempt_id AND lease_token=p_lease_token
      AND recovery_set_id=p_recovery_set_id
      AND lifecycle_state='active' AND lease_expires_at>v_now;
    RETURN FOUND;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_claim_next_recovery_set_v2(uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_heartbeat_recovery_execution_v1(uuid,uuid,uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_claim_next_recovery_set_v2(uuid,uuid) TO forgecrud_backup_reader;
GRANT EXECUTE ON FUNCTION platform.platformops_heartbeat_recovery_execution_v1(uuid,uuid,uuid) TO forgecrud_backup_reader;
