-- +goose Up
GRANT EXECUTE ON FUNCTION iam_v1_actor_assertion_authority_v1(
    name,uuid,uuid,bigint,bigint,bigint
) TO harpy_owner;

