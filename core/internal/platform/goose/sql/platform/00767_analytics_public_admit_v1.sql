-- +goose Up
-- BELIRSIZ: admission does not determine a lease id.
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_public_admit_v1(
    p_tenant_id uuid,
    p_release_id uuid,
    p_generation bigint,
    p_trusted_client_ip inet,
    p_occurred_at timestamptz
)
RETURNS uuid
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT '00000000-0000-0000-0000-000000000000'::uuid;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.analytics_public_admit_v1(uuid, uuid, bigint, inet, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.analytics_public_admit_v1(uuid, uuid, bigint, inet, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.analytics_public_admit_v1(uuid, uuid, bigint, inet, timestamptz);
