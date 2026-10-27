-- +goose Up
ALTER TABLE documents_archives
    ADD COLUMN request_document bytea NOT NULL
        CHECK (pg_catalog.octet_length(request_document) BETWEEN 2 AND 65536),
    ADD COLUMN artifact_plaintext_bytes bigint
        CHECK (artifact_plaintext_bytes IS NULL OR artifact_plaintext_bytes > 0),
    ADD COLUMN ciphertext_sha256 bytea
        CHECK (ciphertext_sha256 IS NULL OR pg_catalog.octet_length(ciphertext_sha256) = 32),
    ADD COLUMN cipher_format text
        CHECK (cipher_format IS NULL OR cipher_format = 'FCDOC1'),
    ADD CONSTRAINT documents_archives_cipher_ready_check CHECK (
        (lifecycle_state = 'PENDING' AND
         ((artifact_plaintext_bytes IS NULL AND ciphertext_sha256 IS NULL AND
           cipher_format IS NULL) OR
          (artifact_plaintext_bytes IS NOT NULL AND ciphertext_sha256 IS NOT NULL AND
           cipher_format = 'FCDOC1')))
        OR
        (lifecycle_state <> 'PENDING' AND artifact_plaintext_bytes IS NOT NULL AND
         ciphertext_sha256 IS NOT NULL
         AND cipher_format = 'FCDOC1')
    );

GRANT INSERT (request_document) ON documents_archives TO forgecrud_runtime;
GRANT UPDATE (artifact_plaintext_bytes, ciphertext_sha256, cipher_format)
ON documents_archives TO forgecrud_runtime;
