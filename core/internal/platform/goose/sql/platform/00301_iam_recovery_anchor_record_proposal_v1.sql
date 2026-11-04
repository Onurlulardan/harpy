-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_anchor_record_proposal_v1(
    in_operation_id uuid,
    in_tenant_id uuid,
    in_actor_id uuid,
    in_registry_binding_id uuid,
    in_generation bigint,
    in_state_sha256 bytea,
    in_proposal_sha256 bytea,
    in_now timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_record_proposal_v1(uuid, uuid, uuid, uuid, bigint, bytea, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_record_proposal_v1(uuid, uuid, uuid, uuid, bigint, bytea, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_recovery_anchor_record_proposal_v1(uuid, uuid, uuid, uuid, bigint, bytea, bytea, timestamptz);
