-- +goose Up
CREATE TABLE iam_profile_field_revisions (
    schema_version_id uuid NOT NULL,
    resource_kind text NOT NULL,
    field_id uuid NOT NULL,
    validation_schema jsonb NOT NULL,
    validation_sha256 bytea NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_profile_field_revisions_pkey
        PRIMARY KEY (schema_version_id, field_id),
    CONSTRAINT iam_profile_field_revisions_schema_version_fk
        FOREIGN KEY (resource_kind, schema_version_id)
        REFERENCES iam_profile_schema_versions (resource_kind, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_profile_field_revisions_field_fk
        FOREIGN KEY (resource_kind, field_id)
        REFERENCES iam_profile_fields (resource_kind, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_profile_field_revisions_resource_kind_check
        CHECK (resource_kind IN ('user', 'organization', 'role')),
    CONSTRAINT iam_profile_field_revisions_validation_schema_check
        CHECK (pg_catalog.jsonb_typeof(validation_schema) = 'object'),
    CONSTRAINT iam_profile_field_revisions_validation_sha256_check
        CHECK (pg_catalog.octet_length(validation_sha256) = 32)
);
