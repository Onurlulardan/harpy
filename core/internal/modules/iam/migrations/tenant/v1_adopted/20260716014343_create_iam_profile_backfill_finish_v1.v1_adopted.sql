-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_profile_backfill_finish_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_plan_id uuid, p_page_number integer, p_scanned_count bigint,
    p_converted_count bigint, p_source_sha256 bytea, p_result_sha256 bytea
)
RETURNS TABLE (
    successor_job_id uuid, successor_run_at timestamptz,
    successor_reference text, successor_sha256 bytea,
    successor_dedupe_key text, successor_priority integer
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_plan_id uuid;
    v_resource_kind text;
    v_source_schema_id uuid;
    v_target_schema_id uuid;
    v_page_count integer;
    v_source_count bigint;
    v_plan_target_sha256 bytea;
    v_page_job_id uuid;
    v_expected_count integer;
    v_page_source_sha256 bytea;
    v_page_target_sha256 bytea;
    v_next_job_id uuid;
    v_next_run_at timestamptz;
    v_next_reference text;
    v_next_sha256 bytea;
    v_checkpoint_plan_id uuid;
    v_completed_pages integer;
    v_converted_total bigint;
    v_error_total bigint;
    v_rolling_sha256 bytea;
    v_row_errors bigint;
    v_errors bigint;
    v_rolling bytea;
    v_terminal boolean := false;
BEGIN
    IF p_scanned_count < 0 OR p_converted_count < 0 OR
       p_converted_count > p_scanned_count OR
       pg_catalog.octet_length(p_source_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_result_sha256) IS DISTINCT FROM 32 OR
       NOT iam_maintenance_claim_assert_v1(
           p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
           'iam.profile.backfill', p_payload_revision,
           p_payload_reference, p_payload_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile backfill completion is invalid';
    END IF;
    SELECT plan.id, plan.resource_kind, plan.source_schema_id,
           plan.target_schema_id, plan.page_count, plan.source_count,
           plan.expected_target_sha256
    INTO v_plan_id, v_resource_kind, v_source_schema_id,
         v_target_schema_id, v_page_count, v_source_count,
         v_plan_target_sha256
    FROM iam_profile_backfill_plans AS plan
    WHERE plan.id = p_plan_id AND plan.state IN ('planned', 'running')
    FOR UPDATE;
    SELECT page.job_id, page.expected_count, page.source_sha256,
           page.expected_target_sha256
    INTO v_page_job_id, v_expected_count, v_page_source_sha256,
         v_page_target_sha256
    FROM iam_profile_backfill_pages AS page
    WHERE page.plan_id = p_plan_id AND page.page_number = p_page_number
      AND page.job_id = p_job_id AND page.state = 'queued'
    FOR UPDATE;
    SELECT checkpoint.plan_id, checkpoint.completed_pages,
           checkpoint.converted_count, checkpoint.error_count,
           checkpoint.rolling_sha256
    INTO v_checkpoint_plan_id, v_completed_pages, v_converted_total,
         v_error_total, v_rolling_sha256
    FROM iam_profile_backfill_checkpoints AS checkpoint
    WHERE checkpoint.plan_id = p_plan_id FOR UPDATE;
    SELECT pg_catalog.count(*) INTO v_row_errors
    FROM iam_profile_backfill_errors AS failure
    WHERE failure.plan_id = p_plan_id AND failure.page_number = p_page_number;
    IF v_plan_id IS NULL OR v_page_job_id IS NULL OR
       v_checkpoint_plan_id IS NULL OR
       v_completed_pages + 1 <> p_page_number OR
       v_expected_count <> p_scanned_count OR
       v_page_source_sha256 IS DISTINCT FROM p_source_sha256 OR
       p_converted_count + v_row_errors <> p_scanned_count THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile backfill page evidence changed';
    END IF;
    v_errors := v_row_errors;
    IF v_errors = 0 AND
       v_page_target_sha256 IS DISTINCT FROM p_result_sha256 THEN
        v_errors := 1;
    END IF;
    v_rolling := pg_catalog.sha256(
        v_rolling_sha256 || p_result_sha256
    );
    UPDATE iam_profile_backfill_pages
    SET state = CASE WHEN v_errors = 0 THEN 'completed' ELSE 'fenced' END,
        converted_count = p_converted_count, error_count = v_row_errors,
        result_sha256 = p_result_sha256, completed_at = v_now
    WHERE plan_id = p_plan_id AND page_number = p_page_number;
    UPDATE iam_profile_backfill_checkpoints
    SET completed_pages = completed_pages + 1,
        converted_count = converted_count + p_converted_count,
        error_count = error_count + v_errors, rolling_sha256 = v_rolling,
        last_page_number = p_page_number, revision = revision + 1,
        updated_at = v_now
    WHERE plan_id = p_plan_id
    RETURNING completed_pages, converted_count, error_count, rolling_sha256
    INTO v_completed_pages, v_converted_total, v_error_total,
         v_rolling_sha256;
    IF v_errors > 0 THEN
        UPDATE iam_profile_backfill_plans
        SET state = 'fenced', fenced_at = v_now,
            revision = revision + 1, updated_at = v_now
        WHERE id = p_plan_id;
        v_terminal := true;
    ELSE
        SELECT page.job_id, page.scheduled_run_at, page.payload_reference,
               page.payload_sha256
        INTO v_next_job_id, v_next_run_at, v_next_reference, v_next_sha256
        FROM iam_profile_backfill_pages AS page
        WHERE page.plan_id = p_plan_id AND page.page_number = p_page_number + 1
          AND page.state = 'queued';
        IF FOUND THEN
            IF p_page_number >= v_page_count THEN
                RAISE EXCEPTION USING ERRCODE = '23514',
                    MESSAGE = 'IAM profile backfill page count conflicts';
            END IF;
        ELSE
            v_terminal := true;
            IF p_page_number <> v_page_count OR
               v_completed_pages <> v_page_count OR
               v_converted_total <> v_source_count OR
               v_error_total <> 0 OR
               v_rolling_sha256 IS DISTINCT FROM v_plan_target_sha256 THEN
                UPDATE iam_profile_backfill_plans
                SET state = 'fenced', fenced_at = v_now,
                    revision = revision + 1, updated_at = v_now
                WHERE id = p_plan_id;
                v_errors := 1;
                UPDATE iam_profile_backfill_pages SET state = 'fenced'
                WHERE plan_id = p_plan_id AND page_number = p_page_number;
            ELSE
                UPDATE iam_profile_schema_versions SET state = 'retired',
                    retired_at = v_now, updated_at = v_now, revision = revision + 1
                WHERE id = v_source_schema_id
                  AND resource_kind = v_resource_kind AND state = 'active';
                IF NOT FOUND THEN
                    RAISE EXCEPTION USING ERRCODE = '40001',
                        MESSAGE = 'IAM profile source pointer changed';
                END IF;
                UPDATE iam_profile_schema_versions SET state = 'active',
                    activated_at = v_now, updated_at = v_now, revision = revision + 1
                WHERE id = v_target_schema_id
                  AND resource_kind = v_resource_kind AND state = 'draft';
                IF NOT FOUND THEN
                    RAISE EXCEPTION USING ERRCODE = '40001',
                        MESSAGE = 'IAM profile target pointer changed';
                END IF;
                UPDATE iam_profile_backfill_plans SET state = 'activated',
                    activated_at = v_now, revision = revision + 1, updated_at = v_now
                WHERE id = p_plan_id;
            END IF;
        END IF;
    END IF;
    RETURN QUERY SELECT * FROM iam_maintenance_profile_backfill_receipt_v1(
        p_job_id, p_attempt_id, p_payload_revision, p_payload_reference,
        p_payload_sha256, p_scanned_count, p_converted_count, v_errors,
        p_result_sha256, CASE WHEN v_terminal THEN NULL ELSE v_next_job_id END,
        CASE WHEN v_terminal THEN NULL ELSE v_next_run_at END,
        CASE WHEN v_terminal THEN NULL ELSE v_next_reference END,
        CASE WHEN v_terminal THEN NULL ELSE v_next_sha256 END,
        CASE WHEN v_terminal THEN NULL ELSE
            'iam.profile.backfill:' || p_plan_id::text || ':' ||
            (p_page_number + 1)::text END
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_profile_backfill_finish_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,bigint,bigint,bytea,bytea
) FROM PUBLIC;
