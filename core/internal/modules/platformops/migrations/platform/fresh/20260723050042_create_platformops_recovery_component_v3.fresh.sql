-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_record_recovery_component_v3(
    p_recovery_set_id uuid,
    p_attempt_id uuid,
    p_lease_token uuid,
    p_component_kind text,
    p_object_key text,
    p_object_version_id text,
    p_ciphertext_sha256 text,
    p_ciphertext_bytes bigint
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_attempt_state text;
    v_lease_expires_at timestamptz;
    v_recovery_state text;
    v_environment_id uuid;
    v_existing_key text;
    v_existing_version text;
    v_existing_sha256 text;
    v_existing_bytes bigint;
    v_expected_key text;
    v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_backup'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'forgecrud-v2:core-database-backup-executor'
       OR p_component_kind NOT IN (
           'database_archive','documents_inventory','key_custody',
           'owner_anchor','protected_authority'
       )
       OR p_object_version_id IS NULL
       OR pg_catalog.octet_length(p_object_version_id) NOT BETWEEN 1 AND 1024
       OR p_ciphertext_sha256 !~ '^[0-9a-f]{64}$'
       OR p_ciphertext_bytes < 1 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery component record denied';
    END IF;
    SELECT attempt.lifecycle_state,attempt.lease_expires_at
    INTO v_attempt_state,v_lease_expires_at
    FROM platform.platformops_recovery_execution_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id
      AND attempt.recovery_set_id=p_recovery_set_id
      AND attempt.lease_token=p_lease_token
    FOR UPDATE;
    IF NOT FOUND THEN RETURN false; END IF;
    SELECT recovery.state,recovery.environment_id
    INTO v_recovery_state,v_environment_id
    FROM platform.platformops_recovery_sets AS recovery
    WHERE recovery.recovery_set_id=p_recovery_set_id
    FOR UPDATE;
    IF NOT FOUND OR v_attempt_state<>'active'
       OR v_lease_expires_at<=v_now
       OR v_recovery_state NOT IN ('copying','verifying') THEN
        RETURN false;
    END IF;
    v_expected_key := 'environments/' || v_environment_id::text ||
        '/recovery-sets/' || p_recovery_set_id::text ||
        '/components/' || p_component_kind || '.fcbackup1';
    IF p_object_key IS DISTINCT FROM v_expected_key THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery component key denied';
    END IF;
    SELECT object_key,object_version_id,ciphertext_sha256,ciphertext_bytes
    INTO v_existing_key,v_existing_version,v_existing_sha256,v_existing_bytes
    FROM platform.platformops_recovery_components
    WHERE recovery_set_id=p_recovery_set_id
      AND component_kind=p_component_kind;
    IF FOUND THEN
        IF (v_existing_key,v_existing_version,
            v_existing_sha256,v_existing_bytes)
           IS DISTINCT FROM
           (p_object_key,p_object_version_id,
            p_ciphertext_sha256,p_ciphertext_bytes) THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='recovery component conflicts';
        END IF;
    ELSE
        INSERT INTO platform.platformops_recovery_components(
            recovery_set_id,component_kind,object_key,object_version_id,
            ciphertext_sha256,ciphertext_bytes,verified_at
        ) VALUES (
            p_recovery_set_id,p_component_kind,p_object_key,p_object_version_id,
            p_ciphertext_sha256,p_ciphertext_bytes,v_now
        );
    END IF;
    UPDATE platform.platformops_recovery_execution_attempts
    SET heartbeat_at=v_now,lease_expires_at=v_now+interval '2 minutes'
    WHERE attempt_id=p_attempt_id AND lease_token=p_lease_token
      AND lifecycle_state='active';
    RETURN FOUND;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_record_recovery_component_v3(
    uuid,uuid,uuid,text,text,text,text,bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_record_recovery_component_v3(
    uuid,uuid,uuid,text,text,text,text,bigint
) TO forgecrud_backup_reader;
