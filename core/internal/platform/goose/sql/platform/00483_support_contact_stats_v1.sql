-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION platform.support_contact_stats_v1(
    arg_1 uuid,
    arg_2 bigint,
    arg_3 bigint,
    arg_4 timestamptz
)
RETURNS bytea
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT NULL::bytea;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.support_contact_stats_v1(uuid, bigint, bigint, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.support_contact_stats_v1(uuid, bigint, bigint, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.support_contact_stats_v1(uuid, bigint, bigint, timestamptz);
