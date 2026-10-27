-- +goose Up
REVOKE EXECUTE ON FUNCTION iam_maintenance_profile_lineage_resources_read_v2(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer
) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION iam_maintenance_profile_backfill_apply_v2(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,uuid,uuid,
    bigint,bytea,bytea,bytea
) FROM harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_maintenance_profile_lineage_resources_read_v3(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_profile_backfill_apply_v3(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,uuid,uuid,
    bigint,bytea,bytea,bytea
) TO harpy_runtime;
