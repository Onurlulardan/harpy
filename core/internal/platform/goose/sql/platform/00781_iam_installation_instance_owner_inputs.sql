-- +goose Up
CREATE TABLE platform.iam_installation_instance_owner_inputs (
    operation_id uuid NOT NULL CHECK (
        operation_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000000'::uuid
    ),
    environment_id uuid PRIMARY KEY CHECK (
        environment_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000000'::uuid
    ),
    tenant_id uuid NOT NULL CHECK (
        tenant_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000000'::uuid
    ),
    schema_name name NOT NULL CHECK (schema_name::text <> ''),
    user_id uuid NOT NULL CHECK (
        user_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000000'::uuid
    ),
    subject_id uuid NOT NULL CHECK (
        subject_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000000'::uuid
    ),
    owner_role_id uuid NOT NULL CHECK (
        owner_role_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000000'::uuid
    ),
    role_permission_id uuid NOT NULL CHECK (
        role_permission_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000000'::uuid
    ),
    owner_binding_id uuid NOT NULL CHECK (
        owner_binding_id IS DISTINCT FROM '00000000-0000-0000-0000-000000000000'::uuid
    ),
    authorization_version bigint NOT NULL,
    request_sha256 text NOT NULL CHECK (request_sha256 <> '')
);

REVOKE ALL ON TABLE platform.iam_installation_instance_owner_inputs FROM PUBLIC;

GRANT SELECT, INSERT ON TABLE platform.iam_installation_instance_owner_inputs TO harpy_api, harpy_worker;

-- +goose Down
DROP TABLE platform.iam_installation_instance_owner_inputs;
