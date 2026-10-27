-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_change_audit_write_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_resource_kind text, p_aggregate_id uuid, p_request_sha256 bytea,
    p_command_key text, p_result_code text
)
RETURNS TABLE (
    aggregate_revision bigint, authorization_clock bigint, event_sequence bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_command text; v_request bytea; v_result text; v_result_id uuid;
    v_receipt_revision bigint; v_receipt_clock bigint; v_receipt_created timestamptz;
    v_revision bigint; v_clock bigint; v_sequence bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_resource_kind NOT IN ('user', 'role', 'organization') OR
       p_command_key NOT LIKE 'iam.identity.' || p_resource_kind || '.%' OR
       p_result_code IS DISTINCT FROM p_command_key || '.succeeded' OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM identity change audit scope is invalid';
    END IF;
    SELECT receipt.command_key, receipt.request_sha256, receipt.result_code,
           receipt.result_id, receipt.result_revision, receipt.clock_value,
           receipt.created_at
    INTO v_command, v_request, v_result, v_result_id, v_receipt_revision,
         v_receipt_clock, v_receipt_created
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id;
    IF v_command IS DISTINCT FROM p_command_key OR
       v_request IS DISTINCT FROM p_request_sha256 OR
       v_result IS DISTINCT FROM p_result_code OR
       v_result_id IS DISTINCT FROM p_aggregate_id OR
       v_receipt_created IS DISTINCT FROM pg_catalog.transaction_timestamp() OR
       NOT EXISTS (
           SELECT 1 FROM iam_users AS actor
           WHERE actor.id = p_actor_id AND actor.status = 'active'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM identity change audit evidence is invalid';
    END IF;
    CASE p_resource_kind
        WHEN 'user' THEN
            SELECT revision INTO v_revision FROM iam_users
            WHERE id = p_aggregate_id FOR UPDATE;
        WHEN 'role' THEN
            SELECT revision INTO v_revision FROM iam_roles
            WHERE id = p_aggregate_id FOR UPDATE;
        WHEN 'organization' THEN
            SELECT revision INTO v_revision FROM iam_organizations
            WHERE id = p_aggregate_id FOR UPDATE;
    END CASE;
    SELECT clock_value INTO v_clock FROM iam_authorization_clock
    WHERE singleton FOR UPDATE;
    IF v_revision IS NULL OR v_clock IS NULL OR
       v_receipt_revision > v_revision OR v_receipt_clock > v_clock THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM identity change result is invalid';
    END IF;
    SELECT COALESCE(pg_catalog.max(audit.event_sequence), 0) + 1
    INTO v_sequence FROM iam_identity_change_audit AS audit
    WHERE audit.resource_kind = p_resource_kind
      AND audit.aggregate_id = p_aggregate_id;
    INSERT INTO iam_identity_change_audit (
        operation_id, actor_id, resource_kind, aggregate_id, request_sha256,
        command_key, result_code, aggregate_revision, authorization_clock,
        event_sequence, recorded_at
    ) VALUES (
        p_operation_id, p_actor_id, p_resource_kind, p_aggregate_id,
        p_request_sha256, p_command_key, p_result_code, v_revision, v_clock,
        v_sequence, pg_catalog.transaction_timestamp()
    );
    RETURN QUERY SELECT v_revision, v_clock, v_sequence;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_change_audit_write_v1(
    name,uuid,uuid,text,uuid,bytea,text,text
) FROM PUBLIC;
