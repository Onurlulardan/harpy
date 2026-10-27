-- +goose Up
GRANT EXECUTE ON FUNCTION iam_profile_schema_publish_v1(
    name,uuid,bytea,uuid,text,bigint,jsonb,bytea,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_profile_document_backfill_v1(
    name,uuid,bytea,text,uuid,uuid,uuid,bytea,bytea,bigint,
    uuid,uuid,bigint,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_profile_schema_retire_v1(
    name,uuid,bytea,uuid,text,bigint,jsonb,bytea,uuid,uuid,bigint,
    bytea,integer,uuid,uuid,bigint,bigint
) TO harpy_runtime;
