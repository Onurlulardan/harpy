-- +goose Up
GRANT EXECUTE ON FUNCTION iam_v1_adoption_plan_register_v1(
    name, uuid, uuid, uuid, bytea, text
) TO harpy_adopter;

GRANT EXECUTE ON FUNCTION iam_v1_adoption_apply_and_record_v1(
    uuid, uuid, text, uuid, smallint, text, text, pg_lsn, pg_lsn, pg_lsn,
    pg_lsn, xid8, integer, text, jsonb, bytea, text, bigint, text
) TO harpy_adopter;

GRANT EXECUTE ON FUNCTION iam_v1_adoption_project_v1(
    name, uuid, uuid
) TO harpy_adopter;

GRANT EXECUTE ON FUNCTION iam_v1_adoption_verify_v1(
    name, uuid, uuid
) TO harpy_adopter;

