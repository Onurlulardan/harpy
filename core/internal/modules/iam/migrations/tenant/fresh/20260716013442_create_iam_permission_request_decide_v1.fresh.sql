-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_decide_v1(
    p_scope_schema name,
    p_operation_id uuid,
    p_request_sha256 bytea,
    p_group_id uuid,
    p_expected_revision bigint,
    p_expected_state text,
    p_reviewer_user_id uuid,
    p_expected_reviewer_token_version bigint,
    p_expected_reviewer_authorization_version bigint,
    p_review_task_id uuid,
    p_decision text,
    p_step_up_purpose_sha256 bytea,
    p_step_up_auth_time timestamptz,
    p_step_up_acr text,
    p_amr_password boolean,
    p_amr_webauthn boolean,
    p_amr_totp boolean
)
RETURNS TABLE (
    result_code text,
    result_revision bigint,
    authorization_clock bigint,
    replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_group_state text;
    v_group_revision bigint;
    v_group_expires_at timestamptz;
    v_group_reviewer_id uuid;
    v_group_task_id uuid;
    v_group_step_up_sha256 bytea;
    v_reviewer_permission_id uuid;
    v_reviewer_state text;
    v_reviewer_token_version bigint;
    v_reviewer_authorization_version bigint;
    v_receipt_command text;
    v_receipt_request_sha256 bytea;
    v_receipt_code text;
    v_receipt_revision bigint;
    v_receipt_clock bigint;
    v_receipt_result_id uuid;
    v_request_id uuid;
    v_binding_id uuid;
    v_bumped_users integer;
    v_clock bigint;
    v_result_code text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission request decision scope is invalid';
    END IF;
    IF p_operation_id IS NULL OR p_group_id IS NULL OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_expected_revision < 1 OR p_expected_state IS DISTINCT FROM 'IN_REVIEW' OR
       p_reviewer_user_id IS NULL OR p_review_task_id IS NULL OR
       p_expected_reviewer_token_version < 1 OR
       p_expected_reviewer_authorization_version < 1 OR
       p_decision NOT IN ('APPROVE', 'REJECT') OR
       pg_catalog.octet_length(p_step_up_purpose_sha256) IS DISTINCT FROM 32 OR
       p_step_up_acr IS DISTINCT FROM 'urn:forgecrud:acr:step-up' OR
       NOT (p_amr_password OR p_amr_webauthn OR p_amr_totp) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM permission request decision input is invalid';
    END IF;
    SELECT request_group.state, request_group.revision,
           request_group.expires_at, request_group.reviewer_user_id,
           request_group.review_task_id, request_group.step_up_purpose_sha256,
           request_group.reviewer_authority_permission_id
    INTO v_group_state, v_group_revision, v_group_expires_at,
         v_group_reviewer_id, v_group_task_id, v_group_step_up_sha256,
         v_reviewer_permission_id
    FROM iam_permission_request_groups AS request_group
    WHERE request_group.id OPERATOR(pg_catalog.=) p_group_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM permission request group is unknown';
    END IF;
    SELECT reviewer.status, reviewer.token_version,
           reviewer.authorization_version
    INTO v_reviewer_state, v_reviewer_token_version,
         v_reviewer_authorization_version
    FROM iam_users AS reviewer
    WHERE reviewer.id OPERATOR(pg_catalog.=) p_reviewer_user_id
    FOR UPDATE;
    IF NOT FOUND OR v_reviewer_state IS DISTINCT FROM 'active' OR
       v_reviewer_token_version IS DISTINCT FROM p_expected_reviewer_token_version OR
       v_reviewer_authorization_version IS DISTINCT FROM
           p_expected_reviewer_authorization_version OR
       v_group_reviewer_id IS DISTINCT FROM p_reviewer_user_id OR
       v_group_task_id IS DISTINCT FROM p_review_task_id OR
       v_group_step_up_sha256 IS DISTINCT FROM p_step_up_purpose_sha256 OR
       p_step_up_auth_time IS NULL OR p_step_up_auth_time > v_now OR
       p_step_up_auth_time < v_now - INTERVAL '10 minutes' OR
       NOT iam_user_has_permission_v1(
           p_reviewer_user_id, v_reviewer_permission_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission request decision evidence is denied';
    END IF;
    SELECT receipt.command_key, receipt.request_sha256, receipt.result_code,
           receipt.result_revision, receipt.clock_value, receipt.result_id
    INTO v_receipt_command, v_receipt_request_sha256, v_receipt_code,
         v_receipt_revision, v_receipt_clock, v_receipt_result_id
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id OPERATOR(pg_catalog.=) p_operation_id;
    IF FOUND THEN
        IF v_receipt_command IS DISTINCT FROM
               'iam.permission_request.group_decide' OR
           v_receipt_request_sha256 IS DISTINCT FROM p_request_sha256 OR
           v_receipt_result_id IS DISTINCT FROM p_group_id THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM permission request operation identity was reused';
        END IF;
        RETURN QUERY SELECT v_receipt_code, v_receipt_revision,
            v_receipt_clock, true;
        RETURN;
    END IF;
    IF v_group_state IS DISTINCT FROM p_expected_state OR
       v_group_revision IS DISTINCT FROM p_expected_revision OR
       v_now >= v_group_expires_at THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM permission request decision CAS is stale';
    END IF;
    PERFORM request.id FROM iam_permission_requests AS request
    WHERE request.group_id = p_group_id
    ORDER BY request.id FOR UPDATE;
    IF p_decision = 'APPROVE' THEN
        FOR v_request_id IN
            SELECT request.id FROM iam_permission_requests AS request
            WHERE request.group_id = p_group_id ORDER BY request.ordinal
        LOOP
            v_binding_id := iam_permission_request_grant_v1(v_request_id, v_now);
        END LOOP;
        v_bumped_users := iam_permission_request_bump_users_v1(p_group_id, v_now);
        v_result_code := 'iam.permission_request.approved';
    ELSE
        v_result_code := 'iam.permission_request.rejected';
    END IF;
    UPDATE iam_authorization_clock
    SET clock_value = clock_value + 1, revision = revision + 1, updated_at = v_now
    WHERE singleton
    RETURNING clock_value INTO v_clock;
    IF v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM authorization clock is unavailable';
    END IF;
    UPDATE iam_permission_request_groups
    SET state = CASE WHEN p_decision = 'APPROVE' THEN 'APPROVED' ELSE 'REJECTED' END,
        terminal_at = v_now, revision = v_group_revision + 1, updated_at = v_now
    WHERE id = p_group_id;
    UPDATE iam_permission_requests AS request
    SET state = CASE WHEN p_decision = 'APPROVE' THEN 'APPROVED' ELSE 'REJECTED' END,
        granted_binding_id = CASE
            WHEN p_decision = 'APPROVE' THEN planned_binding_id ELSE NULL END,
        terminal_at = v_now, revision = revision + 1, updated_at = v_now
    WHERE request.group_id = p_group_id;
    INSERT INTO iam_authorization_operation_receipts (
        operation_id, command_key, request_sha256, clock_value, result_code,
        result_id, result_revision, result_sha256, created_at
    ) VALUES (
        p_operation_id, 'iam.permission_request.group_decide', p_request_sha256,
        v_clock, v_result_code, p_group_id, v_group_revision + 1,
        p_request_sha256, v_now
    );
    RETURN QUERY SELECT v_result_code, v_group_revision + 1, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_decide_v1(
    name,uuid,bytea,uuid,bigint,text,uuid,bigint,bigint,uuid,text,
    bytea,timestamptz,text,boolean,boolean,boolean
) FROM PUBLIC;
