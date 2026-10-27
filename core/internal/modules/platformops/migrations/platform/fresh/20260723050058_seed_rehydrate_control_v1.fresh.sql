-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_seed_rehydrate_control_v1(
    p_trust_sha256 text,p_trust_bytes bytea,
    p_plan_sha256 text,p_plan_bytes bytea,
    p_manifest_sha256 text,p_manifest_bytes bytea
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_trust jsonb;
    v_plan jsonb;
    v_manifest jsonb;
    v_component_key text;
    v_component_value jsonb;
    v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_restore'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'forgecrud-v2:core-database-restore'
       OR p_trust_sha256<>pg_catalog.encode(pg_catalog.sha256(p_trust_bytes),'hex')
       OR p_plan_sha256<>pg_catalog.encode(pg_catalog.sha256(p_plan_bytes),'hex')
       OR p_manifest_sha256<>pg_catalog.encode(
            pg_catalog.sha256(p_manifest_bytes),'hex'
       ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='rehydrate control seed denied';
    END IF;
    v_trust:=pg_catalog.convert_from(p_trust_bytes,'UTF8')::jsonb;
    v_plan:=pg_catalog.convert_from(p_plan_bytes,'UTF8')::jsonb;
    v_manifest:=pg_catalog.convert_from(p_manifest_bytes,'UTF8')::jsonb;
    IF v_trust->>'schema'<>'recovery-trust-bundle.v1'
       OR v_plan->>'schema'<>'forgecrud-recovery-artifact.v1'
       OR v_plan->>'purpose'<>'restore_plan'
       OR v_manifest->>'schema'<>'forgecrud-recovery-artifact.v1'
       OR v_manifest->>'purpose'<>'recovery_manifest'
       OR v_plan->>'environment_id'<>v_trust->>'environment_id'
       OR v_manifest->>'environment_id'<>v_trust->>'environment_id'
       OR v_plan#>>'{payload,identity,environment_id}'<>
          v_trust->>'environment_id'
       OR v_manifest#>>'{payload,identity,environment_id}'<>
          v_trust->>'environment_id'
       OR v_plan#>>'{payload,recovery_set_id}'<>
          v_manifest->>'artifact_id'
       OR v_plan#>>'{payload,manifest_sha256}'<>p_manifest_sha256
       OR v_manifest#>>'{payload,recovery_trust_bundle_sha256}'<>
          p_trust_sha256
       OR (v_manifest->>'key_generation')::bigint<>
          (v_trust->>'generation')::bigint
       OR (v_plan->>'key_generation')::bigint<>
          (v_trust->>'generation')::bigint
       OR v_manifest#>>'{payload,schema}'<>'recovery-set-manifest.v1'
       OR v_plan#>>'{payload,schema}'<>'restore-plan-payload.v1'
       OR v_manifest#>>'{payload,portability}'<>'same_environment'
       OR v_manifest#>>'{payload,identity,deployment_mode}'<>'onpremise'
       OR pg_catalog.jsonb_object_length(
            v_manifest#>'{payload,components}'
          )<>5 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='rehydrate artifacts differ';
    END IF;
    INSERT INTO platform.platformops_recovery_signing_bundles(
        environment_id,generation,bundle_sha256,bundle_bytes,created_at
    ) VALUES (
        (v_trust->>'environment_id')::uuid,
        (v_trust->>'generation')::bigint,p_trust_sha256,p_trust_bytes,
        (v_trust->>'generated_at')::timestamptz
    ) ON CONFLICT DO NOTHING;
    IF NOT EXISTS (
        SELECT 1 FROM platform.platformops_recovery_signing_bundles
        WHERE environment_id=(v_trust->>'environment_id')::uuid
          AND generation=(v_trust->>'generation')::bigint
          AND bundle_sha256=p_trust_sha256 AND bundle_bytes=p_trust_bytes
    ) THEN RAISE EXCEPTION USING ERRCODE='23505',
        MESSAGE='rehydrate trust conflicts'; END IF;
    INSERT INTO platform.platformops_recovery_sets(
        recovery_set_id,operation_id,environment_id,deployment_mode,
        portability,sequence,state,primary_binding_id,tenant_set_sha256,
        owner_binding_id,owner_binding_generation,database_lineage_id,
        database_system_id,recovery_signing_generation,
        recovery_trust_bundle_sha256,manifest_sha256,manifest_bytes,
        protected_at,created_at,updated_at
    ) VALUES (
        (v_manifest->>'artifact_id')::uuid,
        (v_manifest->>'operation_id')::uuid,
        (v_manifest->>'environment_id')::uuid,'onpremise','same_environment',
        (v_manifest->>'artifact_sequence')::bigint,'protected',
        (v_manifest#>>'{payload,identity,primary_binding_id}')::uuid,
        v_manifest#>>'{payload,identity,tenant_set_sha256}',
        (v_manifest#>>'{payload,identity,owner_binding_id}')::uuid,
        (v_manifest#>>'{payload,identity,owner_binding_generation}')::bigint,
        (v_manifest#>>'{payload,identity,database_lineage_id}')::uuid,
        (v_manifest#>>'{payload,database_system_id}')::uuid,
        (v_manifest->>'key_generation')::bigint,p_trust_sha256,
        p_manifest_sha256,p_manifest_bytes,
        (v_manifest#>>'{payload,created_at}')::timestamptz,
        (v_manifest#>>'{payload,created_at}')::timestamptz,v_now
    ) ON CONFLICT DO NOTHING;
    IF NOT EXISTS (
        SELECT 1 FROM platform.platformops_recovery_sets
        WHERE recovery_set_id=(v_manifest->>'artifact_id')::uuid
          AND environment_id=(v_manifest->>'environment_id')::uuid
          AND manifest_sha256=p_manifest_sha256
          AND manifest_bytes=p_manifest_bytes AND state='protected'
    ) THEN RAISE EXCEPTION USING ERRCODE='23505',
        MESSAGE='rehydrate recovery set conflicts'; END IF;
    FOR v_component_key,v_component_value IN
    SELECT component.key,component.value FROM pg_catalog.jsonb_each(
        v_manifest#>'{payload,components}'
    ) AS component LOOP
        IF v_component_key NOT IN (
            'database_archive','documents_inventory','key_custody',
            'owner_anchor','protected_authority'
        ) THEN RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='rehydrate component is invalid'; END IF;
        INSERT INTO platform.platformops_recovery_components(
            recovery_set_id,component_kind,object_key,object_version_id,
            ciphertext_sha256,ciphertext_bytes,verified_at
        ) VALUES (
            (v_manifest->>'artifact_id')::uuid,v_component_key,
            v_component_value->>'object_key',
            v_component_value->>'object_version_id',
            v_component_value->>'ciphertext_sha256',
            (v_component_value->>'ciphertext_bytes')::bigint,v_now
        ) ON CONFLICT DO NOTHING;
    END LOOP;
    IF (SELECT pg_catalog.count(*) FROM platform.platformops_recovery_components
        WHERE recovery_set_id=(v_manifest->>'artifact_id')::uuid)<>5 THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='rehydrate component inventory conflicts';
    END IF;
    INSERT INTO platform.platformops_restore_operations(
        operation_id,recovery_set_id,environment_id,target_disposition,
        state,plan_sha256,plan_bytes,safety_recovery_set_id,
        safety_recovery_set_sha256,waiver_reason,probe_sha256,
        destructive_confirmed,created_at,updated_at
    ) VALUES (
        (v_plan->>'operation_id')::uuid,
        (v_plan#>>'{payload,recovery_set_id}')::uuid,
        (v_plan->>'environment_id')::uuid,
        v_plan#>>'{payload,target_disposition}','plan_ready',
        p_plan_sha256,p_plan_bytes,
        NULLIF(v_plan#>>'{payload,safety_recovery_set_id}','')::uuid,
        NULLIF(v_plan#>>'{payload,safety_recovery_set_sha256}',''),
        NULLIF(v_plan#>>'{payload,waiver_reason}',''),
        NULLIF(v_plan#>>'{payload,probe_sha256}',''),
        (v_plan#>>'{payload,destructive_confirmed}')::boolean,
        (v_plan->>'issued_at')::timestamptz,v_now
    ) ON CONFLICT DO NOTHING;
    RETURN EXISTS (
        SELECT 1 FROM platform.platformops_restore_operations
        WHERE operation_id=(v_plan->>'operation_id')::uuid
          AND recovery_set_id=(v_manifest->>'artifact_id')::uuid
          AND plan_sha256=p_plan_sha256 AND plan_bytes=p_plan_bytes
          AND state IN ('plan_ready','draining','restoring','verifying','switched')
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_seed_rehydrate_control_v1(
    text,bytea,text,bytea,text,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_seed_rehydrate_control_v1(
    text,bytea,text,bytea,text,bytea
) TO forgecrud_restore;
