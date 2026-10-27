-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_security_artifact_purge_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_successor_job_id uuid
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
    v_count bigint;
    v_step bigint;
    v_result bytea;
BEGIN
    v_count := 0;
    IF NOT iam_maintenance_claim_assert_v1(
        p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
        'iam.security-artifact-purge', p_payload_revision,
        p_payload_reference, p_payload_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM security artifact purge claim is invalid';
    END IF;
    DELETE FROM iam_recovery_codes AS code
    WHERE code.state IN ('consumed', 'revoked')
      AND code.purge_after IS NOT NULL AND code.purge_after <= v_now;
    GET DIAGNOSTICS v_step = ROW_COUNT;
    v_count := v_count + v_step;
    DELETE FROM iam_auth_challenges AS challenge
    WHERE challenge.state <> 'pending' AND challenge.purge_after <= v_now;
    GET DIAGNOSTICS v_step = ROW_COUNT;
    v_count := v_count + v_step;
    DELETE FROM iam_one_time_tokens AS token
    WHERE token.state <> 'pending' AND token.purge_after <= v_now;
    GET DIAGNOSTICS v_step = ROW_COUNT;
    v_count := v_count + v_step;
    v_result := pg_catalog.sha256(pg_catalog.convert_to(
        'iam.security-artifact-purge:' || v_count::text, 'UTF8'
    ));
    UPDATE iam_maintenance_checkpoints
    SET cursor_id = NULL, scanned_count = v_count, affected_count = v_count,
        error_count = 0, result_sha256 = v_result, last_job_id = p_job_id,
        revision = revision + 1, updated_at = v_now
    WHERE kind = 'iam.security-artifact-purge';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM security purge checkpoint is missing';
    END IF;
    RETURN QUERY SELECT * FROM iam_maintenance_complete_recurring_v1(
        p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
        'iam.security-artifact-purge', p_payload_revision,
        p_payload_reference, p_payload_sha256, v_count, v_count, 0,
        v_result, p_successor_job_id
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_security_artifact_purge_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid
) FROM PUBLIC;
