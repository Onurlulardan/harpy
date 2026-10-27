-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_record_signing_bundle_v1(
    p_environment_id uuid,
    p_generation bigint,
    p_bundle_sha256 text,
    p_bundle_bytes bytea,
    p_created_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing platform.platformops_recovery_signing_bundles%ROWTYPE;
    v_latest bigint;
BEGIN
    IF p_generation <= 0
       OR p_bundle_sha256 !~ '^[0-9a-f]{64}$'
       OR octet_length(p_bundle_bytes) NOT BETWEEN 1 AND 1048576 THEN
        RAISE EXCEPTION 'invalid recovery signing bundle'
            USING ERRCODE = '23514';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended('platformops:signing:' || p_environment_id::text, 0)
    );
    SELECT * INTO v_existing
    FROM platform.platformops_recovery_signing_bundles
    WHERE environment_id = p_environment_id AND generation = p_generation;
    IF FOUND THEN
        IF v_existing.bundle_sha256 <> p_bundle_sha256
           OR v_existing.bundle_bytes <> p_bundle_bytes
           OR v_existing.created_at <> p_created_at THEN
            RAISE EXCEPTION 'recovery signing generation conflicts'
                USING ERRCODE = '23505';
        END IF;
        RETURN;
    END IF;
    SELECT max(generation) INTO v_latest
    FROM platform.platformops_recovery_signing_bundles
    WHERE environment_id = p_environment_id;
    IF p_generation <> COALESCE(v_latest, 0) + 1 THEN
        RAISE EXCEPTION 'recovery signing generation is not monotonic'
            USING ERRCODE = '40001';
    END IF;
    INSERT INTO platform.platformops_recovery_signing_bundles(
        environment_id, generation, bundle_sha256, bundle_bytes, created_at
    ) VALUES (
        p_environment_id, p_generation, p_bundle_sha256, p_bundle_bytes, p_created_at
    );
END;
$function$;

CREATE FUNCTION platform.platformops_request_recovery_set_v1(
    p_recovery_set_id uuid,
    p_operation_id uuid,
    p_environment_id uuid,
    p_primary_binding_id uuid,
    p_tenant_set_sha256 text,
    p_owner_binding_id uuid,
    p_owner_binding_generation bigint,
    p_database_lineage_id uuid,
    p_recovery_signing_generation bigint,
    p_recovery_trust_bundle_sha256 text,
    p_created_at timestamptz
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing platform.platformops_recovery_sets%ROWTYPE;
    v_sequence bigint;
BEGIN
    IF p_tenant_set_sha256 !~ '^[0-9a-f]{64}$'
       OR p_recovery_trust_bundle_sha256 !~ '^[0-9a-f]{64}$'
       OR p_owner_binding_generation <= 0
       OR p_recovery_signing_generation <= 0 THEN
        RAISE EXCEPTION 'invalid recovery set request'
            USING ERRCODE = '23514';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM platform.platformops_recovery_signing_bundles
        WHERE environment_id = p_environment_id
          AND generation = p_recovery_signing_generation
          AND bundle_sha256 = p_recovery_trust_bundle_sha256
    ) THEN
        RAISE EXCEPTION 'recovery signing bundle is missing'
            USING ERRCODE = 'P0002';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended('platformops:recovery:' || p_environment_id::text, 0)
    );
    SELECT * INTO v_existing
    FROM platform.platformops_recovery_sets
    WHERE operation_id = p_operation_id OR recovery_set_id = p_recovery_set_id
    ORDER BY operation_id = p_operation_id DESC
    LIMIT 1;
    IF FOUND THEN
        IF v_existing.recovery_set_id <> p_recovery_set_id
           OR v_existing.operation_id <> p_operation_id
           OR v_existing.environment_id <> p_environment_id
           OR v_existing.primary_binding_id <> p_primary_binding_id
           OR v_existing.tenant_set_sha256 <> p_tenant_set_sha256
           OR v_existing.owner_binding_id <> p_owner_binding_id
           OR v_existing.owner_binding_generation <> p_owner_binding_generation
           OR v_existing.database_lineage_id <> p_database_lineage_id
           OR v_existing.recovery_signing_generation <> p_recovery_signing_generation
           OR v_existing.recovery_trust_bundle_sha256 <> p_recovery_trust_bundle_sha256 THEN
            RAISE EXCEPTION 'recovery set request conflicts'
                USING ERRCODE = '23505';
        END IF;
        RETURN v_existing.sequence;
    END IF;
    SELECT COALESCE(max(sequence), 0) + 1 INTO v_sequence
    FROM platform.platformops_recovery_sets
    WHERE environment_id = p_environment_id;
    INSERT INTO platform.platformops_recovery_sets(
        recovery_set_id, operation_id, environment_id, deployment_mode,
        portability, sequence, state, primary_binding_id, tenant_set_sha256,
        owner_binding_id, owner_binding_generation, database_lineage_id,
        recovery_signing_generation, recovery_trust_bundle_sha256,
        created_at, updated_at
    ) VALUES (
        p_recovery_set_id, p_operation_id, p_environment_id, 'onpremise',
        'same_environment', v_sequence, 'requested', p_primary_binding_id,
        p_tenant_set_sha256, p_owner_binding_id, p_owner_binding_generation,
        p_database_lineage_id, p_recovery_signing_generation,
        p_recovery_trust_bundle_sha256, p_created_at, p_created_at
    );
    RETURN v_sequence;
END;
$function$;
-- +goose StatementEnd
