-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_reconcile_restored_identity_v1(
    in_restore_operation_id uuid,
    in_environment_id uuid,
    in_recovery_set_id uuid,
    in_owner_user_id uuid,
    in_anchor_id uuid,
    in_security_epoch bytea,
    in_restore_receipt_sha256 bytea,
    in_protected_head_sha256 bytea,
    in_custody_generation bigint,
    in_custody_result_sha256 bytea,
    in_occurred_at timestamptz
)
RETURNS TABLE (
    invalidation_counts jsonb,
    invalidation_sha256 bytea,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '{}'::jsonb,
        ''::bytea,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_reconcile_restored_identity_v1(uuid, uuid, uuid, uuid, uuid, bytea, bytea, bytea, bigint, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_reconcile_restored_identity_v1(uuid, uuid, uuid, uuid, uuid, bytea, bytea, bytea, bigint, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_reconcile_restored_identity_v1(uuid, uuid, uuid, uuid, uuid, bytea, bytea, bytea, bigint, bytea, timestamptz);
