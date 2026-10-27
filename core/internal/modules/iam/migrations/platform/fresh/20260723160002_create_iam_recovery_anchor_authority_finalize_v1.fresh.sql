-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_anchor_finalize_authority_v1(
    p_operation_id uuid
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_operation platform.iam_recovery_anchor_operations%ROWTYPE;
BEGIN
    IF p_operation_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='recovery anchor authority finalization is invalid';
    END IF;
    SELECT * INTO v_operation
      FROM platform.iam_recovery_anchor_operations AS operation
     WHERE operation.operation_id=p_operation_id
     FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P0002',
            MESSAGE='recovery anchor operation is missing';
    END IF;
    IF v_operation.state='completed' THEN
        RETURN true;
    END IF;
    PERFORM result.operation_id
      FROM platform.iam_recovery_anchor_commit_v1(
        v_operation.operation_id,v_operation.tenant_id,
        v_operation.actor_user_id,pg_catalog.clock_timestamp()
      ) AS result;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_finalize_authority_v1(uuid)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_finalize_authority_v1(uuid)
TO forgecrud_runtime;
