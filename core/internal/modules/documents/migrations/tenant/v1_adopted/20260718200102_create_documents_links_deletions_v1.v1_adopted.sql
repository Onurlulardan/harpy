-- +goose Up
CREATE TABLE documents_links (
    id uuid PRIMARY KEY,
    item_id uuid NOT NULL,
    source_owner text NOT NULL CHECK (source_owner IN (
        'FORMS','WORKFLOW','WORK','CATALOG','IAM'
    )),
    source_kind text NOT NULL CHECK (pg_catalog.octet_length(source_kind) BETWEEN 1 AND 128),
    source_id uuid NOT NULL,
    source_version_id uuid NOT NULL,
    usage text NOT NULL CHECK (usage IN ('READ','LIFECYCLE')),
    reference_epoch bigint NOT NULL CHECK (reference_epoch > 0),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('ACTIVE','RELEASED')),
    created_at timestamptz NOT NULL,
    released_at timestamptz,
    CONSTRAINT documents_links_item_fk FOREIGN KEY (item_id)
        REFERENCES documents_items(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    UNIQUE (item_id, source_owner, source_kind, source_id, source_version_id, usage),
    CHECK ((lifecycle_state = 'ACTIVE' AND released_at IS NULL) OR
           (lifecycle_state = 'RELEASED' AND released_at IS NOT NULL))
);

CREATE TABLE documents_legal_holds (
    id uuid PRIMARY KEY,
    item_id uuid NOT NULL,
    source_owner text NOT NULL CHECK (source_owner IN ('BILLING','FORMS','CATALOG','IAM')),
    source_case_id text NOT NULL
        CHECK (pg_catalog.octet_length(source_case_id) BETWEEN 1 AND 256),
    evidence_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(evidence_sha256) = 32),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('ACTIVE','RELEASED')),
    placed_at timestamptz NOT NULL,
    released_at timestamptz,
    CONSTRAINT documents_legal_holds_item_fk FOREIGN KEY (item_id)
        REFERENCES documents_items(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    UNIQUE (item_id, source_owner, source_case_id),
    CHECK ((lifecycle_state = 'ACTIVE' AND released_at IS NULL) OR
           (lifecycle_state = 'RELEASED' AND released_at IS NOT NULL))
);

CREATE TABLE documents_blob_deletions (
    id uuid PRIMARY KEY,
    target_kind text NOT NULL CHECK (target_kind IN ('VERSION','ARCHIVE','UPLOAD_ATTEMPT')),
    target_id uuid NOT NULL,
    object_key text NOT NULL CHECK (pg_catalog.octet_length(object_key) BETWEEN 1 AND 1024),
    object_version text NOT NULL CHECK (pg_catalog.octet_length(object_version) BETWEEN 1 AND 1024),
    object_generation bigint NOT NULL CHECK (object_generation > 0),
    plaintext_bytes bigint NOT NULL CHECK (plaintext_bytes >= 0),
    ciphertext_bytes bigint NOT NULL CHECK (ciphertext_bytes > 0),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN (
        'PENDING','CLAIMED','RECONCILING','CONFIRMED'
    )),
    lease_owner_id uuid,
    lease_generation bigint NOT NULL CHECK (lease_generation >= 0),
    lease_expires_at timestamptz,
    provider_request_id text,
    delete_response_sha256 bytea CHECK (
        delete_response_sha256 IS NULL OR pg_catalog.octet_length(delete_response_sha256) = 32
    ),
    absence_probe_sha256 bytea CHECK (
        absence_probe_sha256 IS NULL OR pg_catalog.octet_length(absence_probe_sha256) = 32
    ),
    billing_adjustment_receipt_id uuid,
    confirmed_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (target_kind, target_id, object_generation),
    CHECK ((lifecycle_state = 'CLAIMED') =
           (lease_owner_id IS NOT NULL AND lease_expires_at IS NOT NULL)),
    CHECK ((lifecycle_state = 'CONFIRMED') =
           (delete_response_sha256 IS NOT NULL AND absence_probe_sha256 IS NOT NULL AND
            billing_adjustment_receipt_id IS NOT NULL AND confirmed_at IS NOT NULL))
);

CREATE INDEX documents_links_source_idx
    ON documents_links(source_owner, source_kind, source_id, source_version_id)
    WHERE lifecycle_state = 'ACTIVE';
CREATE INDEX documents_legal_holds_active_idx
    ON documents_legal_holds(item_id) WHERE lifecycle_state = 'ACTIVE';
CREATE INDEX documents_blob_deletions_claim_idx
    ON documents_blob_deletions(lifecycle_state, lease_expires_at, id)
    WHERE lifecycle_state <> 'CONFIRMED';

