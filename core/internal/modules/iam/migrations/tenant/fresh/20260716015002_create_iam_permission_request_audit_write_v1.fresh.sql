-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_audit_write_v1(
    p_scope_schema name,
    p_operation_id uuid,
    p_request_sha256 bytea,
    p_group_id uuid,
    p_reviewer_user_id uuid,
    p_review_task_id uuid,
    p_decision text,
    p_result_code text,
    p_group_revision bigint,
    p_authorization_clock bigint
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_receipt_command text;
    v_receipt_request bytea;
    v_receipt_result text;
    v_receipt_group uuid;
    v_receipt_revision bigint;
    v_receipt_clock bigint;
    v_group_state text;
    v_group_reviewer uuid;
    v_group_task uuid;
    v_reviewer_active boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission request audit scope is invalid';
    END IF;
    SELECT receipt.command_key, receipt.request_sha256,
           receipt.result_code, receipt.result_id,
           receipt.result_revision, receipt.clock_value
    INTO v_receipt_command, v_receipt_request, v_receipt_result,
         v_receipt_group, v_receipt_revision, v_receipt_clock
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id;
    SELECT request_group.state, request_group.reviewer_user_id,
           request_group.review_task_id
    INTO v_group_state, v_group_reviewer, v_group_task
    FROM iam_permission_request_groups AS request_group
    WHERE request_group.id = p_group_id;
    SELECT true INTO v_reviewer_active
    FROM iam_users AS reviewer
    WHERE reviewer.id = p_reviewer_user_id AND reviewer.status = 'active';
    IF v_receipt_command IS DISTINCT FROM
           'iam.permission_request.group_decide' OR
       v_receipt_request IS DISTINCT FROM p_request_sha256 OR
       v_receipt_result IS DISTINCT FROM p_result_code OR
       v_receipt_group IS DISTINCT FROM p_group_id OR
       v_receipt_revision IS DISTINCT FROM p_group_revision OR
       v_receipt_clock IS DISTINCT FROM p_authorization_clock OR
       v_group_reviewer IS DISTINCT FROM p_reviewer_user_id OR
       v_group_task IS DISTINCT FROM p_review_task_id OR
       v_reviewer_active IS DISTINCT FROM true OR
       (p_decision = 'APPROVE' AND
           v_group_state IS DISTINCT FROM 'APPROVED') OR
       (p_decision = 'REJECT' AND
           v_group_state IS DISTINCT FROM 'REJECTED') OR
       p_decision NOT IN ('APPROVE', 'REJECT') THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM permission request audit evidence is invalid';
    END IF;
    INSERT INTO iam_permission_request_decision_audit (
        operation_id, group_id, reviewer_user_id, review_task_id,
        decision, result_code, request_sha256, group_revision,
        authorization_clock, recorded_at
    ) VALUES (
        p_operation_id, p_group_id, p_reviewer_user_id, p_review_task_id,
        p_decision, p_result_code, p_request_sha256, p_group_revision,
        p_authorization_clock, pg_catalog.transaction_timestamp()
    );
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_audit_write_v1(
    name,uuid,bytea,uuid,uuid,uuid,text,text,bigint,bigint
) FROM PUBLIC;
