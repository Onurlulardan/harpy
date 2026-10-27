-- +goose Up
REVOKE EXECUTE ON FUNCTION platform.iam_v1_adopt_protected_authority_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,text
) FROM harpy_adopter;
