-- +goose Up

GRANT EXECUTE ON FUNCTION platform.analytics_legacy_public_resolve_v1(
    text,text,text,text
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.analytics_legacy_public_record_hit_v1(
    text,text,text,text,uuid,uuid,bytea,uuid,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.analytics_legacy_public_ack_consumers_v1(
    text,text,text,text,uuid,uuid,bytea,uuid,bytea,timestamptz
) TO harpy_cluster_admin;
GRANT EXECUTE ON FUNCTION platform.analytics_legacy_public_retire_v1(
    text,text,text,text,uuid,uuid,bytea,uuid,bytea,timestamptz,timestamptz
) TO harpy_cluster_admin;
