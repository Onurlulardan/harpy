-- +goose Up
GRANT EXECUTE ON FUNCTION iam_maintenance_bootstrap_v1(
    name,uuid,text,uuid,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_receipt_read_v1(
    name,uuid,uuid,text,integer,text,bytea
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_bcrypt_retire_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_identifier_retire_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_directory_tombstone_purge_v1(
    name,uuid,uuid,uuid,integer,text,bytea,bigint,bigint,uuid
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_security_artifact_purge_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_directory_repair_apply_v1(
    name,uuid,uuid,uuid,integer,text,bytea,
    uuid,uuid,text,text,bigint,bytea,jsonb
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_directory_repair_finish_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,bigint,bigint,bytea,uuid
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_directory_checksum_finish_v1(
    name,uuid,uuid,uuid,integer,text,bytea,bigint,bigint,bytea,jsonb,uuid
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_profile_backfill_apply_v1(
    name,uuid,uuid,uuid,integer,text,bytea,
    uuid,integer,uuid,bigint,bytea,bytea,bytea
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_profile_backfill_error_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer,uuid,text,bytea
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_profile_backfill_finish_v1(
    name,uuid,uuid,uuid,integer,text,bytea,
    uuid,integer,bigint,bigint,bytea,bytea
) TO forgecrud_runtime;
