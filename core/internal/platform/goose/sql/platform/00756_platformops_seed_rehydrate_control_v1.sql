-- +goose Up
CREATE FUNCTION platform.platformops_seed_rehydrate_control_v1(
    in_trust_sha256 text,
    in_trust_bytes bytea,
    in_plan_sha256 text,
    in_plan_bytes bytea,
    in_manifest_sha256 text,
    in_manifest_bytes bytea
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformops_seed_rehydrate_control_v1(text, bytea, text, bytea, text, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_seed_rehydrate_control_v1(text, bytea, text, bytea, text, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_seed_rehydrate_control_v1(text, bytea, text, bytea, text, bytea);
