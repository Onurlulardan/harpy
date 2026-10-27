-- +goose Up
REVOKE ALL ON FUNCTION iam_identity_change_audit_write_v1(
    name,uuid,uuid,text,uuid,bytea,text,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_identity_change_audit_write_v1(
    name,uuid,uuid,text,uuid,bytea,text,text
) TO harpy_runtime;
