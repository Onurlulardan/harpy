-- +goose Up
GRANT EXECUTE ON FUNCTION platform.iam_snapshot_domain_generations_v1(
    uuid,text,text
) TO harpy_runtime;
