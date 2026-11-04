-- +goose Up
CREATE FUNCTION platform.platformops_record_recovery_components_v1(
    in_recovery_set_id text,
    in_attempt_id text,
    in_lease_token text,
    in_arg_4 jsonb
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformops_record_recovery_components_v1(text, text, text, jsonb) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_record_recovery_components_v1(text, text, text, jsonb) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_record_recovery_components_v1(text, text, text, jsonb);
