-- +goose Up
CREATE TABLE platform.tenant_adoption_issues (
    issue_id uuid PRIMARY KEY,
    tenant_id uuid,
    schema_name text,
    issue_code text NOT NULL CHECK (issue_code IN (
        'missing_schema',
        'extra_schema',
        'schema_collision',
        'sentinel_missing',
        'sentinel_mismatch',
        'checkpoint_mismatch'
    )),
    evidence_sha256 text NOT NULL UNIQUE
        CHECK (evidence_sha256 ~ '^[0-9a-f]{64}$'),
    state text NOT NULL DEFAULT 'FENCED' CHECK (state = 'FENCED'),
    detected_at timestamptz NOT NULL DEFAULT statement_timestamp(),
    CONSTRAINT tenancy_adoption_issue_identity_check
        CHECK (tenant_id IS NOT NULL OR schema_name IS NOT NULL)
);
