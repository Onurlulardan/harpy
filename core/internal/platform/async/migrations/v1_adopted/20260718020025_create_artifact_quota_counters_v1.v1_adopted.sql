-- +goose Up
CREATE TABLE platform.artifact_quota_counters (
    tenant_id uuid PRIMARY KEY,
    used_ciphertext_bytes bigint NOT NULL DEFAULT 0,
    reserved_ciphertext_bytes bigint NOT NULL DEFAULT 0,
    fenced boolean NOT NULL DEFAULT false,
    fence_code text,
    revision bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    CONSTRAINT artifact_quota_counters_tenant_fk FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id) ON DELETE RESTRICT,
    CONSTRAINT artifact_quota_counters_bytes_check CHECK (
        used_ciphertext_bytes BETWEEN 0 AND 21474836480 AND
        reserved_ciphertext_bytes BETWEEN 0 AND 21474836480 AND
        used_ciphertext_bytes + reserved_ciphertext_bytes <= 21474836480
    ),
    CONSTRAINT artifact_quota_counters_fence_check CHECK (
        fenced = (fence_code IS NOT NULL) AND
        (fence_code IS NULL OR fence_code ~ '^[A-Z][A-Z0-9_]{1,63}$')
    ),
    CONSTRAINT artifact_quota_counters_revision_check CHECK (revision >= 0)
);

