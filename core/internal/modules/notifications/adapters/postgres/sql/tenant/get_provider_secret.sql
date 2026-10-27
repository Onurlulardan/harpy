SELECT id::text, ciphertext, nonce, wrapped_dek, kek_revision, plaintext_digest
FROM notifications_provider_secrets
WHERE id = $1::uuid;
