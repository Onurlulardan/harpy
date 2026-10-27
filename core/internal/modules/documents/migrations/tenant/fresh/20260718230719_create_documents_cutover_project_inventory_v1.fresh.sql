-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_v1_adoption_project_inventory_v1(
    p_operation_id uuid,p_cutover_run_id uuid,p_scope_id uuid,p_record jsonb
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_adoption_run_id uuid;
    v_generation bigint;
    v_manifest text;
    v_created_at timestamptz;
BEGIN
    v_adoption_run_id:=(p_record->>'adoption_run_id')::uuid;
    v_created_at:=(p_record->>'created_at')::timestamptz;
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Documents V1 inventory projection principal denied';
    END IF;
    SELECT operation.cutover_generation,operation.lease_manifest_sha256
    INTO STRICT v_generation,v_manifest
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id
      AND operation.run_id=p_cutover_run_id
      AND operation.scope_kind='tenant' AND operation.scope_id=p_scope_id
      AND operation.group_id='documents' AND operation.domain='documents'
      AND operation.state IN ('copying','shadow')
      AND operation.lease_expires_at>pg_catalog.clock_timestamp()
    FOR SHARE;
    IF (p_record->>'cutover_generation')::bigint IS DISTINCT FROM v_generation OR
       p_record->>'manifest_sha256' IS DISTINCT FROM v_manifest OR
       p_record->>'transform_version' IS DISTINCT FROM 'documents_adoption_v1' OR
       v_created_at>pg_catalog.transaction_timestamp() THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 inventory authority differs';
    END IF;
    INSERT INTO documents_adoption_runs (
        run_id,cutover_generation,manifest_sha256,transform_version,
        lifecycle_state,source_count,target_count,source_plaintext_bytes,
        target_plaintext_bytes,pending_object_count,pending_job_count,
        created_at,updated_at
    ) VALUES (
        v_adoption_run_id,v_generation,pg_catalog.decode(v_manifest,'hex'),
        'documents_adoption_v1','INVENTORY',0,0,0,0,0,0,
        v_created_at,v_created_at
    ) ON CONFLICT (run_id) DO NOTHING;
    IF NOT EXISTS (
        SELECT 1 FROM documents_adoption_runs AS run
        WHERE run.run_id=v_adoption_run_id
          AND run.cutover_generation=v_generation
          AND run.manifest_sha256=pg_catalog.decode(v_manifest,'hex')
          AND run.transform_version='documents_adoption_v1'
          AND run.created_at=v_created_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 inventory replay differs';
    END IF;
    RETURN v_adoption_run_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_v1_adoption_project_inventory_v1(
    uuid,uuid,uuid,jsonb
) FROM PUBLIC;
