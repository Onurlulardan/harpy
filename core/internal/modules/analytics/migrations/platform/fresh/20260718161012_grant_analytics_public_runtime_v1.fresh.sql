-- +goose Up

GRANT EXECUTE ON FUNCTION platform.analytics_public_resolve_v1(bytea) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.analytics_public_admit_v1(
    uuid,uuid,bigint,inet,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.analytics_public_release_v1(
    uuid,uuid,uuid,bigint
) TO forgecrud_runtime;
