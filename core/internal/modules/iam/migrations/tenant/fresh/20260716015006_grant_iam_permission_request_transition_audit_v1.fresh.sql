-- +goose Up
GRANT SELECT ON TABLE iam_permission_request_transition_receipts,
    iam_permission_request_transition_audit
    TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_permission_request_transition_audit_write_v1(
    name,uuid,bytea,uuid,uuid,uuid,text,text,text,bigint
) TO forgecrud_runtime;
