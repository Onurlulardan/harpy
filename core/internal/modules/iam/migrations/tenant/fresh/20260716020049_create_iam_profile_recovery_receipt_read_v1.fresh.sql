-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_recovery_receipt_read_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_manage_permission_id uuid, p_expected_token_version bigint,
    p_expected_authorization_version bigint
)
RETURNS TABLE (
    request_sha256 bytea, plan_id uuid, recovery_of_plan_id uuid,
    recovery_root_plan_id uuid, recovery_sequence bigint,
    result_revision bigint, result_state text, authorization_clock bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_operation_id IS NULL OR NOT iam_profile_actor_assert_v1(
           p_scope_schema,p_actor_id,p_manage_permission_id,
           p_expected_token_version,p_expected_authorization_version
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile recovery receipt authority is denied';
    END IF;
    RETURN QUERY
    SELECT receipt.request_sha256,plan.id,plan.recovery_of_plan_id,
           plan.recovery_root_plan_id,plan.recovery_sequence,
           plan.revision,plan.state,receipt.clock_value
    FROM iam_authorization_operation_receipts AS receipt
    JOIN iam_profile_backfill_plans AS plan ON plan.id = receipt.result_id
    WHERE receipt.operation_id = p_operation_id
      AND receipt.command_key = 'iam.profile_backfill.recovery_create'
      AND plan.recovery_of_plan_id IS NOT NULL
      AND receipt.result_sha256 = plan.plan_sha256;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_recovery_receipt_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
