-- +goose Up
REVOKE EXECUTE ON FUNCTION platform.tenancy_v1_adopt_environment_v1(
    uuid,uuid,text,uuid,text,text
) FROM harpy_adopter;
