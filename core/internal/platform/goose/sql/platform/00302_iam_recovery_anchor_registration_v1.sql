-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_anchor_registration_v1(
    in_id uuid,
    in_tenant_id uuid,
    in_actor_id uuid
)
RETURNS TABLE (
    operation_id bytea,
    anchor_id bytea,
    challenge text,
    rp_id text,
    allowed_origins bytea,
    allowed_origins_sha256 bytea,
    expires_at timestamptz,
    state text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::bytea,
        ''::bytea,
        ''::text,
        ''::text,
        ''::bytea,
        ''::bytea,
        pg_catalog.clock_timestamp(),
        ''::text
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_registration_v1(uuid, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_registration_v1(uuid, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_recovery_anchor_registration_v1(uuid, uuid, uuid);
