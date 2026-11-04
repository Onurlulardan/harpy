-- +goose Up
CREATE FUNCTION platform.iam_post_restore_reconcile_v1(
    in_restore_operation_id uuid,
    in_environment_id uuid,
    in_recovery_set_id uuid,
    in_tenant_id uuid,
    in_owner_user_id uuid,
    in_anchor_id uuid,
    in_new_database_system_id uuid,
    in_security_epoch bytea,
    in_restore_receipt_sha256 bytea,
    in_protected_head_sha256 bytea,
    in_custody_generation bigint,
    in_custody_result_sha256 bytea,
    in_receipt bytea,
    in_occurred_at timestamptz
)
RETURNS TABLE (
    state text,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        false
    ;
$$;

REVOKE ALL ON FUNCTION platform.iam_post_restore_reconcile_v1(uuid, uuid, uuid, uuid, uuid, uuid, uuid, bytea, bytea, bytea, bigint, bytea, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_post_restore_reconcile_v1(uuid, uuid, uuid, uuid, uuid, uuid, uuid, bytea, bytea, bytea, bigint, bytea, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_post_restore_reconcile_v1(uuid, uuid, uuid, uuid, uuid, uuid, uuid, bytea, bytea, bytea, bigint, bytea, bytea, timestamptz);
