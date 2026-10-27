-- +goose Up
CREATE TABLE IF NOT EXISTS platform.platformops_recovery_signing_bundles (
    environment_id uuid NOT NULL,
    generation bigint NOT NULL CHECK (generation > 0),
    bundle_sha256 text NOT NULL CHECK (bundle_sha256 ~ '^[0-9a-f]{64}$'),
    bundle_bytes bytea NOT NULL CHECK (octet_length(bundle_bytes) BETWEEN 1 AND 1048576),
    created_at timestamptz NOT NULL,
    PRIMARY KEY (environment_id, generation),
    UNIQUE (environment_id, bundle_sha256)
);

CREATE TABLE IF NOT EXISTS platform.platformops_recovery_sets (
    recovery_set_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL UNIQUE,
    environment_id uuid NOT NULL,
    deployment_mode text NOT NULL CHECK (deployment_mode = 'onpremise'),
    portability text NOT NULL CHECK (portability = 'same_environment'),
    sequence bigint NOT NULL CHECK (sequence > 0),
    state text NOT NULL CHECK (state IN (
        'requested','holding','snapshotting','copying','verifying',
        'protected','failed','deleting','deleted'
    )),
    primary_binding_id uuid NOT NULL,
    tenant_set_sha256 text NOT NULL CHECK (tenant_set_sha256 ~ '^[0-9a-f]{64}$'),
    owner_binding_id uuid NOT NULL,
    owner_binding_generation bigint NOT NULL CHECK (owner_binding_generation > 0),
    database_lineage_id uuid NOT NULL,
    database_system_id uuid,
    recovery_signing_generation bigint NOT NULL CHECK (recovery_signing_generation > 0),
    recovery_trust_bundle_sha256 text NOT NULL
        CHECK (recovery_trust_bundle_sha256 ~ '^[0-9a-f]{64}$'),
    manifest_sha256 text CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    manifest_bytes bytea CHECK (
        manifest_bytes IS NULL OR octet_length(manifest_bytes) BETWEEN 1 AND 2097152
    ),
    legal_hold boolean NOT NULL DEFAULT false,
    failure_code text,
    protected_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (environment_id, sequence),
    CHECK ((manifest_sha256 IS NULL) = (manifest_bytes IS NULL)),
    CHECK ((state = 'protected') = (protected_at IS NOT NULL)),
    CHECK (state <> 'protected' OR manifest_sha256 IS NOT NULL)
);

CREATE TABLE IF NOT EXISTS platform.platformops_recovery_components (
    recovery_set_id uuid NOT NULL
        REFERENCES platform.platformops_recovery_sets(recovery_set_id),
    component_kind text NOT NULL CHECK (component_kind IN (
        'database_archive','documents_inventory','key_custody',
        'owner_anchor','protected_authority'
    )),
    object_key text NOT NULL CHECK (length(object_key) BETWEEN 1 AND 1024),
    object_version_id text NOT NULL CHECK (length(object_version_id) BETWEEN 1 AND 1024),
    ciphertext_sha256 text NOT NULL CHECK (ciphertext_sha256 ~ '^[0-9a-f]{64}$'),
    ciphertext_bytes bigint NOT NULL CHECK (ciphertext_bytes > 0),
    verified_at timestamptz NOT NULL,
    PRIMARY KEY (recovery_set_id, component_kind),
    UNIQUE (recovery_set_id, object_key, object_version_id)
);

CREATE TABLE IF NOT EXISTS platform.platformops_restore_operations (
    operation_id uuid PRIMARY KEY,
    recovery_set_id uuid NOT NULL
        REFERENCES platform.platformops_recovery_sets(recovery_set_id),
    environment_id uuid NOT NULL,
    target_disposition text NOT NULL CHECK (
        target_disposition IN ('existing_readable','empty','unreadable')
    ),
    state text NOT NULL CHECK (state IN (
        'plan_ready','draining','restoring','verifying','switched',
        'completed','failed','fenced'
    )),
    plan_sha256 text NOT NULL CHECK (plan_sha256 ~ '^[0-9a-f]{64}$'),
    plan_bytes bytea NOT NULL CHECK (octet_length(plan_bytes) BETWEEN 1 AND 2097152),
    safety_recovery_set_id uuid,
    safety_recovery_set_sha256 text
        CHECK (safety_recovery_set_sha256 ~ '^[0-9a-f]{64}$'),
    waiver_reason text CHECK (
        waiver_reason IS NULL OR waiver_reason IN (
            'empty_no_database_or_policy','unreadable_integrity_probe'
        )
    ),
    probe_sha256 text CHECK (probe_sha256 ~ '^[0-9a-f]{64}$'),
    destructive_confirmed boolean NOT NULL DEFAULT false,
    seeded boolean NOT NULL DEFAULT false CHECK (seeded = false),
    owner_changed boolean NOT NULL DEFAULT false CHECK (owner_changed = false),
    tenant_replaced boolean NOT NULL DEFAULT false CHECK (tenant_replaced = false),
    receipt_sha256 text CHECK (receipt_sha256 ~ '^[0-9a-f]{64}$'),
    receipt_bytes bytea CHECK (
        receipt_bytes IS NULL OR octet_length(receipt_bytes) BETWEEN 1 AND 2097152
    ),
    new_database_system_id uuid,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    completed_at timestamptz,
    CHECK ((receipt_sha256 IS NULL) = (receipt_bytes IS NULL)),
    CHECK (
        (target_disposition = 'existing_readable'
            AND safety_recovery_set_id IS NOT NULL
            AND safety_recovery_set_sha256 IS NOT NULL
            AND waiver_reason IS NULL AND probe_sha256 IS NULL
            AND destructive_confirmed = false)
        OR
        (target_disposition = 'empty'
            AND safety_recovery_set_id IS NULL
            AND safety_recovery_set_sha256 IS NULL
            AND waiver_reason = 'empty_no_database_or_policy'
            AND probe_sha256 IS NOT NULL
            AND destructive_confirmed = false)
        OR
        (target_disposition = 'unreadable'
            AND safety_recovery_set_id IS NULL
            AND safety_recovery_set_sha256 IS NULL
            AND waiver_reason = 'unreadable_integrity_probe'
            AND probe_sha256 IS NOT NULL
            AND destructive_confirmed = true)
    )
);
