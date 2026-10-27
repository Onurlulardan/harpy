-- +goose Up

GRANT EXECUTE ON FUNCTION platform.analytics_public_resolve_v1(bytea) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.analytics_public_admit_v1(
    uuid,uuid,bigint,inet,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.analytics_public_release_v1(
    uuid,uuid,uuid,bigint
) TO harpy_runtime;
