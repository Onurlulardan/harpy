SELECT artifact.id, artifact.subject_user_id, artifact.resource_id,
       artifact.token_id, artifact.purpose, artifact.artifact_ref,
       recipient.normalized_value, artifact.kms_key_id, artifact.nonce,
       artifact.ciphertext, artifact.authentication_tag, artifact.aad_sha256,
       artifact.ciphertext_sha256, artifact.expires_at
FROM iam_lifecycle_delivery_artifacts AS artifact
JOIN iam_one_time_tokens AS token ON token.id = artifact.token_id
JOIN iam_user_identifiers AS recipient
  ON recipient.id = artifact.recipient_identifier_id
WHERE artifact.artifact_ref = $1 AND token.state = 'pending'
  AND token.expires_at > statement_timestamp()
  AND artifact.expires_at > statement_timestamp()
  AND ((artifact.purpose = 'email_verification_v1' AND
        recipient.state IN ('pending', 'verified')) OR
       (artifact.purpose IN (
            'password_reset_v1', 'portal_setup_v1', 'initial_admin_setup.v1'
        ) AND
        recipient.state = 'active'))
