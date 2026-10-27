-- +goose Up
GRANT EXECUTE ON FUNCTION platform.platformops_restored_license_scope_mint_v1(
    uuid,uuid,uuid,text,uuid,uuid,text,text,bigint,text,bigint,text,text,bigint
) TO harpy_runtime;
