-- +goose Up
GRANT EXECUTE ON FUNCTION iam_authorization_evidence_read_v2(
    name,uuid,uuid,text,text
) TO harpy_runtime;
