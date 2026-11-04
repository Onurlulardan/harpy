-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_field_reference_evidence_v2(
    arg_1 name,
    arg_2 text,
    arg_3 uuid
)
RETURNS TABLE (
    live_reference_count bigint,
    manifest_sha256 bytea,
    checked_owner_count integer
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        0::bigint,
        NULL::bytea,
        0;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_field_reference_evidence_v2(name, text, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_field_reference_evidence_v2(name, text, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION system_resource_layout_field_reference_evidence_v2(name, text, uuid);
