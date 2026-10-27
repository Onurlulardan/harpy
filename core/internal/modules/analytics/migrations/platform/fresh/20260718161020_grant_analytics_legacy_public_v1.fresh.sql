-- +goose Up

GRANT EXECUTE ON FUNCTION platform.analytics_legacy_public_resolve_v1(
    text,text,text,text
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.analytics_legacy_public_record_hit_v1(
    text,text,text,text,uuid,uuid,bytea,uuid,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.analytics_legacy_public_ack_consumers_v1(
    text,text,text,text,uuid,uuid,bytea,uuid,bytea,timestamptz
) TO forgecrud_cluster_admin;
GRANT EXECUTE ON FUNCTION platform.analytics_legacy_public_retire_v1(
    text,text,text,text,uuid,uuid,bytea,uuid,bytea,timestamptz,timestamptz
) TO forgecrud_cluster_admin;
