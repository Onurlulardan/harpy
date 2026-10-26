UPDATE notifications_provider_secrets
SET wrapped_dek = $2, kek_revision = $3,
    rewrapped_at = pg_catalog.transaction_timestamp()
WHERE id = $1::uuid AND kek_revision = $4
  AND ciphertext = $5 AND nonce = $6 AND plaintext_digest = $7;
