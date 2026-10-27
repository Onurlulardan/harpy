-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_bcrypt_retire_v1(
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
    v_cutover_at timestamptz;
    v_count bigint;
    v_result bytea;
BEGIN
    v_count := 0;
    IF NOT iam_maintenance_claim_assert_v1(
        p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
        'iam.credential.bcrypt-retire', p_payload_revision,
        p_payload_reference, p_payload_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM bcrypt retirement claim is invalid';
    END IF;
    SELECT cutover.changed_at INTO v_cutover_at
    FROM platform.domain_cutovers AS cutover
    WHERE cutover.scope_kind = 'tenant' AND cutover.scope_id = p_tenant_id
      AND cutover.domain = 'iam' AND cutover.state = 'v2'
      AND cutover.writer = 'v2'
    FOR SHARE;
    IF v_cutover_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM bcrypt retirement cutover is unavailable';
    END IF;
    IF v_now >= v_cutover_at + INTERVAL '180 days' THEN
        UPDATE iam_password_credentials AS credential
        SET state = 'reset_required', algorithm = 'legacy_bcrypt_retired_v1',
            legacy_bcrypt_hash = NULL, legacy_retired_by_job_id = p_job_id,
            revision = revision + 1, updated_at = v_now
        WHERE credential.algorithm = 'bcrypt_v1_adopted';
        GET DIAGNOSTICS v_count = ROW_COUNT;
        UPDATE iam_refresh_tokens AS token
        SET state = 'revoked', revoked_at = v_now,
            revision = revision + 1, updated_at = v_now
        WHERE token.state = 'active' AND EXISTS (
            SELECT 1 FROM iam_refresh_families AS family
            JOIN iam_password_credentials AS credential
              ON credential.user_id = family.user_id
            WHERE family.id = token.family_id
              AND credential.legacy_retired_by_job_id = p_job_id
        );
        UPDATE iam_refresh_families AS family
        SET state = 'revoked', revoke_reason = 'bcrypt_retired',
            revoked_at = v_now, revision = revision + 1, updated_at = v_now
        WHERE family.state = 'active' AND EXISTS (
            SELECT 1 FROM iam_password_credentials AS credential
            WHERE credential.user_id = family.user_id
              AND credential.legacy_retired_by_job_id = p_job_id
        );
        UPDATE iam_sessions AS session
        SET state = 'revoked', revoked_at = v_now,
            revision = revision + 1, updated_at = v_now
        WHERE session.state = 'active' AND EXISTS (
            SELECT 1 FROM iam_password_credentials AS credential
            WHERE credential.user_id = session.user_id
              AND credential.legacy_retired_by_job_id = p_job_id
        );
        UPDATE iam_users AS user_record
        SET token_version = token_version + 1, revision = revision + 1,
            updated_at = v_now
        WHERE EXISTS (
            SELECT 1 FROM iam_password_credentials AS credential
            WHERE credential.user_id = user_record.id
              AND credential.legacy_retired_by_job_id = p_job_id
        );
    END IF;
    v_result := pg_catalog.sha256(pg_catalog.convert_to(
        'iam.credential.bcrypt-retire:' || v_count::text, 'UTF8'
    ));
    RETURN QUERY SELECT * FROM iam_maintenance_complete_recurring_v1(
        p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
        'iam.credential.bcrypt-retire', p_payload_revision,
        p_payload_reference, p_payload_sha256, v_count, v_count, 0,
        v_result, p_successor_job_id
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_bcrypt_retire_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid
) FROM PUBLIC;
