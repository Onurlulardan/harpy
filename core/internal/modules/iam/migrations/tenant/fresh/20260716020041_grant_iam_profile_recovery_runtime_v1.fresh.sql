-- +goose Up

REVOKE EXECUTE ON FUNCTION iam_profile_read_contract_v1(name,text)
    FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION iam_profile_write_contract_read_v1(name,text)
    FROM harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_profile_recovery_inventory_open_v1(
    name,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_profile_read_contract_v2(name,text)
    TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_profile_write_contract_read_v2(name,text)
    TO harpy_runtime;
