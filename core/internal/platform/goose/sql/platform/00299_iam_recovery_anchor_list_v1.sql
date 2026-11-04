-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_anchor_list_v1(
    in_environment_id uuid,
    in_tenant_id uuid,
    in_actor_id uuid
)
RETURNS TABLE (
    anchor_id bytea,
    generation bigint,
    credential_id bytea,
    rp_id text,
    user_verification text,
    backup_eligible boolean,
    backup_state boolean,
    state text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::bytea,
        0::bigint,
        ''::bytea,
        ''::text,
        ''::text,
        false,
        false,
        ''::text
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_list_v1(uuid, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_list_v1(uuid, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_recovery_anchor_list_v1(uuid, uuid, uuid);
