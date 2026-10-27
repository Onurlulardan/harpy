-- +goose Up
GRANT SELECT (
    receipt_id, artifact_ref, effect_idempotency_key, message_id,
    outcome, code, provider_reference, result_sha256, completed_at
) ON iam_lifecycle_delivery_worker_receipts TO harpy_runtime;
GRANT INSERT (
    receipt_id, artifact_id, purpose, subject_user_id, resource_id, token_id,
    artifact_ref, effect_id, effect_idempotency_key, message_id, outcome, code,
    provider_reference, result_sha256, completed_at
) ON iam_lifecycle_delivery_worker_receipts TO harpy_runtime;

GRANT SELECT (
    artifact_id, artifact_ref, effect_id, effect_idempotency_key,
    message_id, state, claimed_at, outcome_unknown_at
) ON iam_lifecycle_delivery_worker_claims TO harpy_runtime;
GRANT INSERT (
    artifact_id, artifact_ref, effect_id, effect_idempotency_key,
    message_id, state, claimed_at, outcome_unknown_at
) ON iam_lifecycle_delivery_worker_claims TO harpy_runtime;
GRANT UPDATE (state, outcome_unknown_at)
ON iam_lifecycle_delivery_worker_claims TO harpy_runtime;
GRANT DELETE ON iam_lifecycle_delivery_worker_claims TO harpy_runtime;

GRANT SELECT (
    id, purpose, subject_user_id, resource_id, recipient_identifier_id,
    token_id, artifact_ref, kms_key_id, nonce, ciphertext,
    authentication_tag, aad_sha256, ciphertext_sha256, expires_at
) ON iam_lifecycle_delivery_artifacts TO harpy_runtime;
GRANT DELETE ON iam_lifecycle_delivery_artifacts TO harpy_runtime;

GRANT SELECT (id, state, expires_at)
ON iam_one_time_tokens TO harpy_runtime;
GRANT SELECT (id, normalized_value, state)
ON iam_user_identifiers TO harpy_runtime;
