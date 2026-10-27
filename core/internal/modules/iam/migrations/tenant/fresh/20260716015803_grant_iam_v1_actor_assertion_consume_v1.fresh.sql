-- +goose Up
GRANT EXECUTE ON FUNCTION iam_v1_actor_assertion_consume_v1(
    name,uuid,uuid,uuid,uuid,text,text,text,bytea,bytea,
    bigint,bigint,bigint,bigint,bigint,bigint,bigint,bigint
) TO forgecrud_runtime;

