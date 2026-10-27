-- +goose Up
REVOKE EXECUTE ON FUNCTION iam_maintenance_profile_lineage_resources_read_v2(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_maintenance_profile_backfill_apply_v2(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,uuid,uuid,
    bigint,bytea,bytea,bytea
) FROM forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_maintenance_profile_lineage_resources_read_v3(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_profile_backfill_apply_v3(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,uuid,uuid,
    bigint,bytea,bytea,bytea
) TO forgecrud_runtime;
