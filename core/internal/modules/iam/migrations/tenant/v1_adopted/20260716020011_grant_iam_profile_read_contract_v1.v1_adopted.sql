-- +goose Up

GRANT EXECUTE ON FUNCTION iam_profile_read_contract_v1(name,text)
    TO harpy_v2_runtime;
