-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_audit_write_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid, p_aggregate_id uuid,
    p_request_sha256 bytea, p_command_key text, p_result_code text,
    p_aggregate_revision bigint, p_authorization_clock bigint
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_command text; v_request bytea; v_result text; v_aggregate uuid;
    v_revision bigint; v_clock bigint; v_actor_active boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'IAM RBAC audit scope is invalid';
    END IF;
    SELECT receipt.command_key, receipt.request_sha256, receipt.result_code,
        receipt.result_id, receipt.result_revision, receipt.clock_value
    INTO v_command, v_request, v_result, v_aggregate, v_revision, v_clock
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id;
    SELECT true INTO v_actor_active FROM iam_users AS actor
    WHERE actor.id = p_actor_id AND actor.status = 'active';
    IF v_command IS DISTINCT FROM p_command_key OR v_request IS DISTINCT FROM p_request_sha256 OR
       v_result IS DISTINCT FROM p_result_code OR v_aggregate IS DISTINCT FROM p_aggregate_id OR
       v_revision IS DISTINCT FROM p_aggregate_revision OR
       v_clock IS DISTINCT FROM p_authorization_clock OR v_actor_active IS DISTINCT FROM true THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'IAM RBAC audit evidence is invalid';
    END IF;
    INSERT INTO iam_rbac_change_audit (
        operation_id, actor_id, aggregate_id, request_sha256, command_key,
        result_code, aggregate_revision, authorization_clock, recorded_at
    ) VALUES (p_operation_id, p_actor_id, p_aggregate_id, p_request_sha256,
        p_command_key, p_result_code, p_aggregate_revision, p_authorization_clock,
        pg_catalog.transaction_timestamp());
    RETURN true;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_rbac_audit_write_v1(
    name,uuid,uuid,uuid,bytea,text,text,bigint,bigint
) FROM PUBLIC;
