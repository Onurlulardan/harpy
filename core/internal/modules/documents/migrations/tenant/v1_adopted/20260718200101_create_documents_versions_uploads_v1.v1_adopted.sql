-- +goose Up
CREATE TABLE documents_versions (
    id uuid PRIMARY KEY,
    item_id uuid NOT NULL,
    version_number bigint NOT NULL CHECK (version_number > 0),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN (
        'READY','DELETE_PENDING','PURGED'
    )),
    original_name text NOT NULL
        CHECK (pg_catalog.octet_length(original_name) BETWEEN 1 AND 255),
    declared_mime text NOT NULL CHECK (pg_catalog.octet_length(declared_mime) <= 255),
    detected_mime text NOT NULL CHECK (pg_catalog.octet_length(detected_mime) <= 255),
    file_kind text NOT NULL CHECK (file_kind IN (
        'PDF','TXT','CSV','JSON','JPEG','PNG','GIF','WEBP','DOC','DOCX',
        'XLS','XLSX','PPT','PPTX','ODT','ODS','ODP'
    )),
    plaintext_bytes bigint NOT NULL CHECK (plaintext_bytes BETWEEN 1 AND 104857600),
    ciphertext_bytes bigint NOT NULL CHECK (ciphertext_bytes > 0),
    plaintext_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(plaintext_sha256) = 32),
    ciphertext_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(ciphertext_sha256) = 32),
    cipher_format text NOT NULL CHECK (cipher_format = 'FCDOC1'),
    object_key text NOT NULL CHECK (pg_catalog.octet_length(object_key) BETWEEN 1 AND 1024),
    object_version text NOT NULL CHECK (pg_catalog.octet_length(object_version) BETWEEN 1 AND 1024),
    object_generation bigint NOT NULL CHECK (object_generation > 0),
    key_id text NOT NULL CHECK (key_id ~ '^[a-z0-9][a-z0-9_.-]{0,63}$'),
    wrapped_dek bytea NOT NULL CHECK (pg_catalog.octet_length(wrapped_dek) BETWEEN 115 AND 256),
    scan_signature_revision text NOT NULL
        CHECK (pg_catalog.octet_length(scan_signature_revision) BETWEEN 1 AND 256),
    scanned_at timestamptz NOT NULL,
    delete_receipt_sha256 bytea CHECK (
        delete_receipt_sha256 IS NULL OR pg_catalog.octet_length(delete_receipt_sha256) = 32
    ),
    purged_at timestamptz,
    created_at timestamptz NOT NULL,
    CONSTRAINT documents_versions_item_fk FOREIGN KEY (item_id)
        REFERENCES documents_items(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT documents_versions_number_key UNIQUE (item_id, version_number),
    CONSTRAINT documents_versions_object_key UNIQUE (object_key, object_version),
    CHECK ((lifecycle_state = 'PURGED' AND delete_receipt_sha256 IS NOT NULL AND
            purged_at IS NOT NULL) OR lifecycle_state <> 'PURGED')
);

ALTER TABLE documents_items ADD CONSTRAINT documents_items_current_version_fk
    FOREIGN KEY (current_version_id) REFERENCES documents_versions(id)
    ON UPDATE NO ACTION ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;

CREATE TABLE documents_uploads (
    id uuid PRIMARY KEY,
    actor_id uuid NOT NULL,
    document_id uuid NOT NULL,
    candidate_version_id uuid NOT NULL UNIQUE,
    expected_current_version_id uuid,
    folder_id uuid NOT NULL,
    request_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(request_sha256) = 32),
    expected_content_sha256 bytea NOT NULL
        CHECK (pg_catalog.octet_length(expected_content_sha256) = 32),
    expected_plaintext_bytes bigint NOT NULL
        CHECK (expected_plaintext_bytes BETWEEN 1 AND 104857600),
    ciphertext_upper_bound bigint NOT NULL CHECK (ciphertext_upper_bound > 0),
    original_name text NOT NULL
        CHECK (pg_catalog.octet_length(original_name) BETWEEN 1 AND 255),
    declared_mime text NOT NULL CHECK (pg_catalog.octet_length(declared_mime) <= 255),
    billing_reservation_id uuid NOT NULL UNIQUE,
    attempt_generation bigint NOT NULL CHECK (attempt_generation >= 0),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN (
        'CREATED','RECEIVING','RETRYABLE','SCANNING','READY',
        'REJECTED','CONFLICTED','CANCELLED','EXPIRED'
    )),
    lease_owner_id uuid,
    lease_expires_at timestamptz,
    cleanup_required boolean NOT NULL,
    cleanup_receipt_sha256 bytea CHECK (
        cleanup_receipt_sha256 IS NULL OR pg_catalog.octet_length(cleanup_receipt_sha256) = 32
    ),
    stable_error_code text,
    ready_status smallint CHECK (ready_status IS NULL OR ready_status = 201),
    ready_response bytea CHECK (
        ready_response IS NULL OR pg_catalog.octet_length(ready_response) <= 65536
    ),
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT documents_uploads_folder_fk FOREIGN KEY (folder_id)
        REFERENCES documents_folders(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK ((lifecycle_state = 'RECEIVING') =
           (lease_owner_id IS NOT NULL AND lease_expires_at IS NOT NULL)),
    CHECK ((lifecycle_state = 'READY') =
           (ready_status IS NOT NULL AND ready_response IS NOT NULL)),
    CHECK (NOT cleanup_required OR lifecycle_state IN (
        'RETRYABLE','REJECTED','CONFLICTED','CANCELLED'
    ))
);

CREATE UNIQUE INDEX documents_uploads_idempotency_key
    ON documents_uploads(actor_id, request_sha256);
CREATE INDEX documents_uploads_reconcile_idx
    ON documents_uploads(expires_at, id)
    WHERE lifecycle_state NOT IN ('READY','CANCELLED','EXPIRED');
