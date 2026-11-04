-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_maintenance_directory_snapshot_v1(
    in_tenant_id uuid
)
RETURNS TABLE (
    source_identifier_id uuid,
    lookup_key_id text,
    identifier_hmac bytea,
    user_id uuid,
    source_revision bigint,
    identifier_kind text,
    state text,
    source_sha256 bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::bytea,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::text,
        ''::text,
        ''::bytea
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_maintenance_directory_snapshot_v1(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_maintenance_directory_snapshot_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_maintenance_directory_snapshot_v1(uuid);
