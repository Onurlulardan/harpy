-- +goose Up
CREATE TABLE platform.cutover_credential_receipts (
    receipt_id uuid NOT NULL,
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    lease_id uuid NOT NULL,
    receipt_kind text NOT NULL,
    contract_version text NOT NULL,
    lease_manifest_sha256 text NOT NULL,
    seal_receipt_sha256 text,
    receipt_sha256 text NOT NULL,
    canonical_document jsonb NOT NULL,
    issued_at timestamptz NOT NULL,
    verified_at timestamptz NOT NULL,
    PRIMARY KEY (receipt_id),
    CONSTRAINT cutover_credential_receipts_operation_fk
        FOREIGN KEY (operation_id, run_id)
        REFERENCES platform.cutover_operations (operation_id, run_id)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_credential_receipts_kind_unique
        UNIQUE (operation_id, run_id, receipt_kind),
    CONSTRAINT cutover_credential_receipts_hash_unique
        UNIQUE (operation_id, run_id, receipt_sha256),
    CONSTRAINT cutover_credential_receipts_seal_fk
        FOREIGN KEY (operation_id, run_id, seal_receipt_sha256)
        REFERENCES platform.cutover_credential_receipts (operation_id, run_id, receipt_sha256)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_credential_receipts_kind_check
        CHECK (receipt_kind IN ('session_seal', 'credential_cleanup')),
    CONSTRAINT cutover_credential_receipts_contract_check
        CHECK (
            (receipt_kind = 'session_seal' AND
                contract_version = 'cutover-session-seal-receipt.v1' AND
                seal_receipt_sha256 IS NULL) OR
            (receipt_kind = 'credential_cleanup' AND
                contract_version = 'cutover-credential-cleanup-receipt.v1' AND
                seal_receipt_sha256 IS NOT NULL)
        ),
    CONSTRAINT cutover_credential_receipts_digest_check
        CHECK (
            lease_manifest_sha256 ~ '^[0-9a-f]{64}$' AND
            receipt_sha256 ~ '^[0-9a-f]{64}$' AND
            (seal_receipt_sha256 IS NULL OR seal_receipt_sha256 ~ '^[0-9a-f]{64}$')
        ),
    CONSTRAINT cutover_credential_receipts_document_check
        CHECK (pg_catalog.octet_length(canonical_document::text) BETWEEN 2 AND 65536),
    CONSTRAINT cutover_credential_receipts_time_check CHECK (issued_at <= verified_at)
);

CREATE INDEX cutover_credential_receipts_lease_idx
    ON platform.cutover_credential_receipts (lease_id, operation_id, run_id);
