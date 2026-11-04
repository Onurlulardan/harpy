-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_anchor_stage_delete_v1(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_actor_id uuid,
    in_anchor_id uuid,
    in_idempotency_key text,
    in_request_sha256 bytea,
    in_occurred_at timestamptz
)
RETURNS TABLE (
    anchor_id bytea,
    anchor_set_sha256 bytea,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::bytea,
        ''::bytea,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_stage_delete_v1(uuid, uuid, uuid, uuid, uuid, text, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_stage_delete_v1(uuid, uuid, uuid, uuid, uuid, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_recovery_anchor_stage_delete_v1(uuid, uuid, uuid, uuid, uuid, text, bytea, timestamptz);
