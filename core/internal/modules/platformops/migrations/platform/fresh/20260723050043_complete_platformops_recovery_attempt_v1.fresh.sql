-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_complete_recovery_execution_v1(
    p_recovery_set_id uuid,
    p_attempt_id uuid,
    p_lease_token uuid,
    p_manifest_sha256 text,
    p_manifest_bytes bytea,
    p_database_system_id uuid
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_attempt_state text;
    v_attempt_phase text;
    v_attempt_lease timestamptz;
    v_documents_hold_sha256 text;
    v_recovery_state text;
    v_stored_manifest_sha256 text;
    v_stored_manifest_bytes bytea;
    v_stored_database_system_id uuid;
    v_environment_id uuid;
    v_operation_id uuid;
    v_sequence bigint;
    v_signing_generation bigint;
    v_primary_binding_id uuid;
    v_tenant_set_sha256 text;
    v_owner_binding_id uuid;
    v_owner_binding_generation bigint;
    v_database_lineage_id uuid;
    v_trust_bundle_sha256 text;
    v_manifest jsonb;
    v_kinds text;
    v_database_sha text;
    v_key_custody_sha text;
    v_owner_anchor_sha text;
    v_authority_sha text;
    v_now timestamptz;
BEGIN
    v_now := pg_catalog.statement_timestamp();
    IF SESSION_USER IS DISTINCT FROM 'harpy_backup'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'forgecrud-v2:core-database-backup-executor'
       OR p_recovery_set_id IS NULL OR p_attempt_id IS NULL
       OR p_lease_token IS NULL OR p_database_system_id IS NULL
       OR p_manifest_sha256 IS NULL OR p_manifest_bytes IS NULL
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$'
       OR pg_catalog.octet_length(p_manifest_bytes) NOT BETWEEN 1 AND 2097152
       OR p_manifest_sha256 <> pg_catalog.encode(
           pg_catalog.sha256(p_manifest_bytes),'hex'
       ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery completion denied';
    END IF;
    SELECT attempt.lifecycle_state,attempt.claimed_recovery_state,
           attempt.lease_expires_at,attempt.documents_hold_sha256
    INTO v_attempt_state,v_attempt_phase,v_attempt_lease,
         v_documents_hold_sha256
    FROM platform.platformops_recovery_execution_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id
      AND attempt.recovery_set_id=p_recovery_set_id
      AND attempt.lease_token=p_lease_token
    FOR UPDATE;
    IF NOT FOUND THEN RETURN false; END IF;
    SELECT recovery.state,recovery.manifest_sha256,recovery.manifest_bytes,
           recovery.database_system_id,recovery.environment_id,
           recovery.operation_id,recovery.sequence,
           recovery.recovery_signing_generation,recovery.primary_binding_id,
           recovery.tenant_set_sha256,recovery.owner_binding_id,
           recovery.owner_binding_generation,recovery.database_lineage_id,
           recovery.recovery_trust_bundle_sha256
    INTO v_recovery_state,v_stored_manifest_sha256,v_stored_manifest_bytes,
         v_stored_database_system_id,v_environment_id,v_operation_id,v_sequence,
         v_signing_generation,v_primary_binding_id,v_tenant_set_sha256,
         v_owner_binding_id,v_owner_binding_generation,v_database_lineage_id,
         v_trust_bundle_sha256
    FROM platform.platformops_recovery_sets AS recovery
    WHERE recovery.recovery_set_id=p_recovery_set_id
    FOR UPDATE;
    IF NOT FOUND THEN RETURN false; END IF;
    IF v_attempt_state='completed' AND v_recovery_state='protected' THEN
        RETURN (v_stored_manifest_sha256,v_stored_manifest_bytes,
                v_stored_database_system_id)
            IS NOT DISTINCT FROM
               (p_manifest_sha256,p_manifest_bytes,p_database_system_id);
    END IF;
    IF v_attempt_state<>'active' OR v_attempt_lease<=v_now
       OR v_attempt_phase<>'verifying'
       OR v_documents_hold_sha256 IS NULL
       OR v_recovery_state<>'verifying' THEN
        RETURN false;
    END IF;
    SELECT pg_catalog.string_agg(component_kind,',' ORDER BY component_kind),
           pg_catalog.max(ciphertext_sha256)
               FILTER (WHERE component_kind='database_archive'),
           pg_catalog.max(ciphertext_sha256)
               FILTER (WHERE component_kind='key_custody'),
           pg_catalog.max(ciphertext_sha256)
               FILTER (WHERE component_kind='owner_anchor'),
           pg_catalog.max(ciphertext_sha256)
               FILTER (WHERE component_kind='protected_authority')
    INTO v_kinds,v_database_sha,v_key_custody_sha,v_owner_anchor_sha,v_authority_sha
    FROM platform.platformops_recovery_components
    WHERE recovery_set_id=p_recovery_set_id;
    IF v_kinds IS DISTINCT FROM
       'database_archive,documents_inventory,key_custody,owner_anchor,protected_authority' THEN
        RETURN false;
    END IF;
    v_manifest := pg_catalog.convert_from(p_manifest_bytes,'UTF8')::jsonb;
    IF v_manifest->>'schema' IS DISTINCT FROM 'forgecrud-recovery-artifact.v1'
       OR v_manifest->>'purpose' IS DISTINCT FROM 'recovery_manifest'
       OR v_manifest->>'environment_id' IS DISTINCT FROM v_environment_id::text
       OR v_manifest->>'operation_id' IS DISTINCT FROM v_operation_id::text
       OR v_manifest->>'artifact_id' IS DISTINCT FROM p_recovery_set_id::text
       OR (v_manifest->>'artifact_sequence')::bigint IS DISTINCT FROM v_sequence
       OR (v_manifest->>'key_generation')::bigint IS DISTINCT FROM
          v_signing_generation
       OR v_manifest->>'algorithm' IS DISTINCT FROM 'Ed25519'
       OR v_manifest->>'payload_sha256' IS NULL
       OR v_manifest->>'payload_sha256' !~ '^[0-9a-f]{64}$'
       OR v_manifest->>'signature' IS NULL
       OR pg_catalog.octet_length(v_manifest->>'signature') < 1
       OR v_manifest->>'issued_at' IS NULL
       OR v_manifest->>'expires_at' IS NULL
       OR (v_manifest->>'issued_at')::timestamptz > v_now
       OR (v_manifest->>'expires_at')::timestamptz <= v_now
       OR v_manifest#>>'{payload,schema}' IS DISTINCT FROM
          'environment-recovery-set-payload.v1'
       OR v_manifest#>>'{payload,portability}' IS DISTINCT FROM 'same_environment'
       OR v_manifest#>>'{payload,identity,environment_id}' IS DISTINCT FROM
          v_environment_id::text
       OR v_manifest#>>'{payload,identity,deployment_mode}' IS DISTINCT FROM 'onpremise'
       OR v_manifest#>>'{payload,identity,primary_binding_id}' IS DISTINCT FROM
          v_primary_binding_id::text
       OR v_manifest#>>'{payload,identity,tenant_set_sha256}' IS DISTINCT FROM
          v_tenant_set_sha256
       OR v_manifest#>>'{payload,identity,owner_binding_id}' IS DISTINCT FROM
          v_owner_binding_id::text
       OR (v_manifest#>>'{payload,identity,owner_binding_generation}')::bigint IS DISTINCT FROM
          v_owner_binding_generation
       OR v_manifest#>>'{payload,identity,database_lineage_id}' IS DISTINCT FROM
          v_database_lineage_id::text
       OR v_manifest#>>'{payload,database_system_id}' IS DISTINCT FROM
          p_database_system_id::text
       OR v_manifest#>>'{payload,database_archive_sha256}' IS DISTINCT FROM v_database_sha
       OR v_manifest#>>'{payload,documents_merkle_sha256}' IS DISTINCT FROM
          v_documents_hold_sha256
       OR v_manifest#>>'{payload,key_custody_sha256}' IS DISTINCT FROM v_key_custody_sha
       OR v_manifest#>>'{payload,owner_anchor_sha256}' IS DISTINCT FROM v_owner_anchor_sha
       OR v_manifest#>>'{payload,protected_head_ack_sha256}' IS DISTINCT FROM v_authority_sha
       OR v_manifest#>>'{payload,recovery_trust_bundle_sha256}' IS DISTINCT FROM
          v_trust_bundle_sha256
       OR (v_manifest#>>'{payload,recovery_signing_generation}')::bigint IS DISTINCT FROM
          v_signing_generation THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery manifest evidence denied';
    END IF;
    UPDATE platform.platformops_recovery_sets
    SET state='protected',manifest_sha256=p_manifest_sha256,
        manifest_bytes=p_manifest_bytes,database_system_id=p_database_system_id,
        protected_at=v_now,updated_at=v_now
    WHERE recovery_set_id=p_recovery_set_id AND state='verifying';
    IF NOT FOUND THEN RETURN false; END IF;
    UPDATE platform.platformops_recovery_execution_attempts
    SET lifecycle_state='completed',heartbeat_at=v_now,
        lease_expires_at=v_now,completed_at=v_now
    WHERE attempt_id=p_attempt_id AND lease_token=p_lease_token
      AND lifecycle_state='active';
    RETURN FOUND;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_complete_recovery_execution_v1(
    uuid,uuid,uuid,text,bytea,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_complete_recovery_execution_v1(
    uuid,uuid,uuid,text,bytea,uuid
) TO harpy_backup_reader;
