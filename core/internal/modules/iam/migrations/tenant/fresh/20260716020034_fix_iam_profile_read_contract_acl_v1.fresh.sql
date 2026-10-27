-- +goose Up

REVOKE EXECUTE ON FUNCTION iam_profile_read_contract_v1(name,text)
    FROM harpy_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_profile_read_contract_v1(name,text)
    TO harpy_runtime;
