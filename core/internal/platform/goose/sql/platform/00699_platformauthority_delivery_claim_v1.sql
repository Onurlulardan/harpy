-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_delivery_claim_v1(
    in_lease_owner_id text
)
RETURNS TABLE (
    operation_id text,
    lease_generation bigint,
    attempt bigint,
    max_attempts bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        0::bigint,
        0::bigint,
        0::bigint;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformauthority_delivery_claim_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_delivery_claim_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_delivery_claim_v1(text);
