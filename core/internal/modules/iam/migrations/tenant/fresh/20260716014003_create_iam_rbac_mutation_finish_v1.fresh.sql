-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_mutation_finish_v1(
    p_operation_id uuid, p_command_key text, p_request_sha256 bytea,
    p_clock bigint, p_result_code text, p_result_id uuid, p_result_revision bigint
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_clock < 1 OR p_result_revision < 1 OR
       p_result_code !~ '^iam\.rbac\.[a-z0-9_.]+$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM RBAC mutation result is invalid';
    END IF;
    INSERT INTO iam_authorization_operation_receipts (
        operation_id, command_key, request_sha256, clock_value,
        result_code, result_id, result_revision, result_sha256, created_at
    ) VALUES (
        p_operation_id, p_command_key, p_request_sha256, p_clock,
        p_result_code, p_result_id, p_result_revision, p_request_sha256,
        pg_catalog.transaction_timestamp()
    );
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_mutation_finish_v1(
    uuid,text,bytea,bigint,text,uuid,bigint
) FROM PUBLIC;
