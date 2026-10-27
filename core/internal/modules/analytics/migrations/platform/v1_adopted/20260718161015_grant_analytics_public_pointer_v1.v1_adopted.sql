-- +goose Up
GRANT EXECUTE ON FUNCTION platform.analytics_public_activate_v1(
    uuid,bytea,uuid,text,uuid,text,bigint,bytea,timestamptz,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.analytics_public_revoke_v1(
    uuid,uuid,bigint,timestamptz
) TO harpy_runtime;
