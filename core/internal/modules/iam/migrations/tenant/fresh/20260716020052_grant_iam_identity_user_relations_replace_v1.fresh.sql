-- +goose Up
GRANT EXECUTE ON FUNCTION iam_identity_user_relations_replace_v1(name,uuid,text,bytea,
    uuid,bigint,boolean,uuid,bytea,bytea,boolean,uuid[],uuid[],boolean,uuid[],uuid[],
    uuid,uuid,bigint,bigint) TO forgecrud_runtime;
