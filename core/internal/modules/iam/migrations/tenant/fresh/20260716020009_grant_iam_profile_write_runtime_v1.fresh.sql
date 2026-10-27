-- +goose Up
GRANT EXECUTE ON FUNCTION iam_profile_write_contract_read_v1(name,text)
    TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_profile_write_resource_read_v1(name,text,uuid)
    TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_profile_preconversion_reserve_v1(
    name,uuid,text,uuid,uuid,bigint,bytea,bytea,bytea
) TO forgecrud_runtime;
