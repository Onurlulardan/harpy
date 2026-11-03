-- +goose Up
CREATE TABLE documents_uploads (
    id uuid NOT NULL PRIMARY KEY,
    actor_id uuid NOT NULL,
    attempt_generation bigint NOT NULL,
    cancel_requested_at timestamptz,
    lifecycle_state text NOT NULL,
    document_id uuid NOT NULL,
    candidate_version_id uuid NOT NULL,
    expected_current_version_id uuid NOT NULL,
    folder_id uuid NOT NULL,
    request_sha256 text NOT NULL,
    expected_content_sha256 text NOT NULL,
    expected_plaintext_bytes bigint NOT NULL,
    ciphertext_upper_bound text NOT NULL,
    original_name text NOT NULL,
    declared_mime text NOT NULL,
    description text NOT NULL,
    tags jsonb NOT NULL,
    billing_reservation_id uuid NOT NULL,
    cleanup_required boolean NOT NULL,
    expires_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    stable_error_code text NOT NULL,
    ready_status text NOT NULL,
    ready_response text NOT NULL,
    row_version bigint NOT NULL,
    lease_expires_at timestamptz
);

-- +goose Down
DROP TABLE documents_uploads;
