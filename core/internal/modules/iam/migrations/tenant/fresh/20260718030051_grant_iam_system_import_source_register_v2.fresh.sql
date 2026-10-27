-- +goose Up

GRANT EXECUTE ON FUNCTION iam_system_import_source_register_v2(
    name,uuid,uuid,uuid,text,bigint,bigint,bytea,text,text,text,bytea,bytea,
    bigint,bytea,bigint,text,bytea,bytea,text,timestamptz
) TO forgecrud_v2_runtime;
