-- +goose Up
REVOKE EXECUTE ON FUNCTION iam_maintenance_profile_schema_read_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,uuid
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_maintenance_profile_resources_read_v2(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_maintenance_profile_backfill_apply_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,uuid,bigint,bytea,bytea,bytea
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_maintenance_profile_backfill_finish_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,bigint,bigint,bytea,bytea
) FROM forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_maintenance_profile_lineage_resources_read_v2(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_profile_schema_read_v2(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,uuid
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_profile_backfill_apply_v2(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,uuid,uuid,
    bigint,bytea,bytea,bytea
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_profile_backfill_finish_v2(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,bigint,bigint,bytea,bytea
) TO forgecrud_runtime;
