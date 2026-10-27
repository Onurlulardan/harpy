-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_v1_adoption_verify_v1(
    p_scope_schema name,
    p_operation_id uuid,
    p_run_id uuid
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_target_table text;
    v_target_record jsonb;
    v_matches boolean;
    v_verified bigint;
    v_state_sha text;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    v_verified := 0;
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT EXISTS (
           SELECT 1 FROM iam_v1_adoption_runs AS run
           WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
             AND run.eligible AND run.state = 'projected'
           FOR UPDATE
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM V1 adoption verification authority or state denied';
    END IF;
    FOR v_target_table, v_target_record IN
        SELECT stage.target_table, stage.target_record
        FROM iam_v1_adoption_stage AS stage
        WHERE stage.operation_id = p_operation_id AND stage.run_id = p_run_id
          AND NOT stage.removed
        ORDER BY stage.target_table, stage.target_key
    LOOP
        v_matches := iam_v1_adoption_identity_record_matches_v1(
            v_target_table, v_target_record
        );
        IF NOT v_matches THEN
            v_matches := iam_v1_adoption_access_record_matches_v1(
                v_target_table, v_target_record
            );
        END IF;
        IF NOT v_matches THEN
            RAISE EXCEPTION USING ERRCODE = '22000',
                MESSAGE = 'IAM V1 adoption target parity differs';
        END IF;
        v_verified := v_verified + 1;
    END LOOP;
    IF v_verified IS DISTINCT FROM (
        SELECT pg_catalog.count(*) FROM iam_v1_adoption_stage AS stage
        WHERE stage.operation_id = p_operation_id AND stage.run_id = p_run_id
          AND NOT stage.removed
    ) OR EXISTS (
        SELECT 1 FROM iam_v1_adoption_category_evidence AS evidence
        WHERE evidence.operation_id = p_operation_id AND evidence.run_id = p_run_id
          AND evidence.category <> 'layouts'
          AND (evidence.copy_state <> 'complete' OR
               evidence.staged_count IS NULL OR evidence.staged_sha256 IS NULL)
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22000',
            MESSAGE = 'IAM V1 adoption category parity is incomplete';
    END IF;
    UPDATE iam_v1_adoption_category_evidence
    SET target_count = staged_count, target_sha256 = staged_sha256, updated_at = v_now
    WHERE operation_id = p_operation_id AND run_id = p_run_id
      AND category <> 'layouts' AND copy_state = 'complete';
    SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        pg_catalog.string_agg(
            evidence.category || ':' || evidence.target_count::text || ':' ||
            pg_catalog.encode(evidence.target_sha256, 'hex'),
            ',' ORDER BY evidence.category
        ), 'UTF8'
    )), 'hex')
    INTO STRICT v_state_sha
    FROM iam_v1_adoption_category_evidence AS evidence
    WHERE evidence.operation_id = p_operation_id AND evidence.run_id = p_run_id
      AND evidence.category <> 'layouts';
    UPDATE iam_v1_adoption_runs
    SET state = 'verified', verified_at = v_now, updated_at = v_now
    WHERE operation_id = p_operation_id AND run_id = p_run_id
      AND state = 'projected';
    IF NOT FOUND OR v_state_sha !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22000',
            MESSAGE = 'IAM V1 adoption verification state raced';
    END IF;
    RETURN v_state_sha;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_v1_adoption_verify_v1(
    name, uuid, uuid
) FROM PUBLIC;

