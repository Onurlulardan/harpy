-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_record_v1_adoption_mapping_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_category text,
    p_source_key text,
    p_record_kind text,
    p_target_id uuid,
    p_record_sha256 bytea,
    p_recorded_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
BEGIN
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id
      AND run.run_id = p_run_id
      AND run.state = 'planned'
      AND run.eligible
    FOR SHARE;
    IF p_category IS NULL OR p_source_key IS NULL
       OR p_record_kind IS NULL OR p_target_id IS NULL
       OR pg_catalog.octet_length(p_record_sha256) <> 32
       OR p_recorded_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing V1 adoption mapping is invalid';
    END IF;
    INSERT INTO platform.billing_v1_adoption_mappings (
        operation_id, run_id, scope_kind, scope_id, authority_id,
        category, source_key, record_kind, target_id,
        record_sha256, created_at
    ) VALUES (
        p_operation_id, p_run_id, v_run.scope_kind, v_run.scope_id,
        v_run.authority_id, p_category, p_source_key, p_record_kind,
        p_target_id, p_record_sha256, p_recorded_at
    );
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_record_v1_adoption_mapping_v1(
    uuid, uuid, text, text, text, uuid, bytea, timestamptz
) FROM PUBLIC;
-- +goose StatementEnd

