-- +goose Up
CREATE TABLE iam_profile_retirement_evidence (
    operation_id uuid NOT NULL,
    resource_kind text NOT NULL,
    field_id uuid NOT NULL,
    field_revision bigint NOT NULL,
    reference_manifest_sha256 bytea NOT NULL,
    checked_owner_count integer NOT NULL,
    live_reference_count bigint NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_profile_retirement_evidence_pkey PRIMARY KEY (operation_id),
    CONSTRAINT iam_profile_retirement_evidence_field_fk
        FOREIGN KEY (resource_kind, field_id)
        REFERENCES iam_profile_fields (resource_kind, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_profile_retirement_evidence_resource_check
        CHECK (resource_kind IN ('user', 'organization', 'role')),
    CONSTRAINT iam_profile_retirement_evidence_revision_check
        CHECK (field_revision > 0),
    CONSTRAINT iam_profile_retirement_evidence_hash_check
        CHECK (pg_catalog.octet_length(reference_manifest_sha256) = 32),
    CONSTRAINT iam_profile_retirement_evidence_owner_check
        CHECK (checked_owner_count > 0),
    CONSTRAINT iam_profile_retirement_evidence_reference_check
        CHECK (live_reference_count = 0)
);

CREATE TRIGGER iam_profile_retirement_evidence_mutation_guard
BEFORE UPDATE OR DELETE OR TRUNCATE ON iam_profile_retirement_evidence
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_profile_destructive_mutation_v1();
ALTER TABLE iam_profile_retirement_evidence
    ENABLE ALWAYS TRIGGER iam_profile_retirement_evidence_mutation_guard;
