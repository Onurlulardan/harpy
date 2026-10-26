INSERT INTO notifications_provider_secrets (
    id, ciphertext, nonce, wrapped_dek, kek_revision, plaintext_digest,
    created_at, rewrapped_at
) VALUES (
    $1::uuid, $2, $3, $4, $5, $6, pg_catalog.transaction_timestamp(), NULL
);
