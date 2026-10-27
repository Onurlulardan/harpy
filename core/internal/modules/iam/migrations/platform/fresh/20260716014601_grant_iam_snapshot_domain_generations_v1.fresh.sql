-- +goose Up
GRANT EXECUTE ON FUNCTION platform.iam_snapshot_domain_generations_v1(
    uuid,text,text
) TO forgecrud_runtime;
