-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_advance_recovery_set_v1(
    p_recovery_set_id uuid,
    p_expected_state text,
    p_next_state text,
    p_failure_code text,
    p_updated_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    UPDATE platform.platformops_recovery_sets
       SET state = p_next_state,
           failure_code = CASE WHEN p_next_state = 'failed' THEN p_failure_code ELSE NULL END,
           updated_at = p_updated_at
     WHERE recovery_set_id = p_recovery_set_id
       AND state = p_expected_state
       AND p_next_state IN ('holding','snapshotting','copying','verifying','failed')
       AND (p_next_state <> 'failed' OR length(p_failure_code) BETWEEN 1 AND 128)
    ;
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION platform.platformops_record_recovery_component_v1(
    p_recovery_set_id uuid,
    p_component_kind text,
    p_object_key text,
    p_object_version_id text,
    p_ciphertext_sha256 text,
    p_ciphertext_bytes bigint,
    p_verified_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing platform.platformops_recovery_components%ROWTYPE;
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM platform.platformops_recovery_sets
         WHERE recovery_set_id = p_recovery_set_id
           AND state IN ('copying','verifying')
    ) THEN
        RAISE EXCEPTION 'recovery set cannot accept components' USING ERRCODE = '55000';
    END IF;
    SELECT * INTO v_existing FROM platform.platformops_recovery_components
     WHERE recovery_set_id = p_recovery_set_id AND component_kind = p_component_kind;
    IF FOUND THEN
        IF v_existing.object_key <> p_object_key
           OR v_existing.object_version_id <> p_object_version_id
           OR v_existing.ciphertext_sha256 <> p_ciphertext_sha256
           OR v_existing.ciphertext_bytes <> p_ciphertext_bytes
           OR v_existing.verified_at <> p_verified_at THEN
            RAISE EXCEPTION 'recovery component conflicts' USING ERRCODE = '23505';
        END IF;
        RETURN;
    END IF;
    INSERT INTO platform.platformops_recovery_components(
        recovery_set_id, component_kind, object_key, object_version_id,
        ciphertext_sha256, ciphertext_bytes, verified_at
    ) VALUES (
        p_recovery_set_id, p_component_kind, p_object_key, p_object_version_id,
        p_ciphertext_sha256, p_ciphertext_bytes, p_verified_at
    );
END;
$function$;

CREATE FUNCTION platform.platformops_complete_recovery_set_v1(
    p_recovery_set_id uuid,
    p_manifest_sha256 text,
    p_manifest_bytes bytea,
    p_database_system_id uuid,
    p_protected_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    UPDATE platform.platformops_recovery_sets
       SET state = 'protected', manifest_sha256 = p_manifest_sha256,
           manifest_bytes = p_manifest_bytes, database_system_id = p_database_system_id,
           protected_at = p_protected_at, updated_at = p_protected_at
     WHERE recovery_set_id = p_recovery_set_id AND state = 'verifying'
       AND p_manifest_sha256 ~ '^[0-9a-f]{64}$'
       AND octet_length(p_manifest_bytes) BETWEEN 1 AND 2097152
       AND (SELECT count(*) FROM platform.platformops_recovery_components c
             WHERE c.recovery_set_id = p_recovery_set_id) = 5
    ;
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION platform.platformops_mark_recovery_set_deleting_v1(
    p_recovery_set_id uuid,
    p_expected_manifest_sha256 text,
    p_updated_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    UPDATE platform.platformops_recovery_sets
       SET state = 'deleting', updated_at = p_updated_at
     WHERE recovery_set_id = p_recovery_set_id AND state = 'protected'
       AND legal_hold = false AND manifest_sha256 = p_expected_manifest_sha256
    ;
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION platform.platformops_finish_recovery_set_delete_v1(
    p_recovery_set_id uuid, p_succeeded boolean, p_updated_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    UPDATE platform.platformops_recovery_sets
       SET state = CASE WHEN p_succeeded THEN 'deleted' ELSE 'protected' END,
           updated_at = p_updated_at
     WHERE recovery_set_id = p_recovery_set_id AND state = 'deleting'
    ;
    RETURN FOUND;
END;
$function$;
-- +goose StatementEnd
