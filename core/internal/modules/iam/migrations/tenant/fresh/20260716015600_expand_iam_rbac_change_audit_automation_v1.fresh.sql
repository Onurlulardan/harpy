-- +goose Up
ALTER TABLE iam_rbac_change_audit
    DROP CONSTRAINT iam_rbac_change_audit_command_check,
    DROP CONSTRAINT iam_rbac_change_audit_result_check;

ALTER TABLE iam_rbac_change_audit
    ADD CONSTRAINT iam_rbac_change_audit_command_check CHECK (
        command_key ~ '^iam\.(rbac|service_principal|capability_grant)\.[a-z0-9_.]+$'
    ),
    ADD CONSTRAINT iam_rbac_change_audit_result_check CHECK (
        result_code ~ '^iam\.(rbac|service_principal|capability_grant)\.[a-z0-9_.]+$'
    );
