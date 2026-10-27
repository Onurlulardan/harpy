-- +goose Up
CREATE TRIGGER iam_auth_challenges_immutable_guard
BEFORE UPDATE ON iam_auth_challenges
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'user_id', 'session_id', 'authenticator_id', 'authenticator_kind',
    'purpose', 'proof_kind', 'hash_profile', 'hmac_key_id', 'challenge_hmac',
    'context_sha256', 'step_up_purpose_sha256', 'origin_policy_version',
    'action_policy_version', 'rp_id_sha256', 'origin_sha256',
    'user_verification_required', 'issued_at', 'expires_at', 'created_at'
);
ALTER TABLE iam_auth_challenges
    ENABLE ALWAYS TRIGGER iam_auth_challenges_immutable_guard;

CREATE TRIGGER iam_invitations_immutable_guard
BEFORE UPDATE ON iam_invitations
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'identifier_kind', 'display_value', 'normalized_value',
    'normalized_sha256', 'role_id', 'role_scope_kind', 'role_scope_id',
    'organization_id', 'invited_by', 'created_at', 'expires_at'
);
ALTER TABLE iam_invitations
    ENABLE ALWAYS TRIGGER iam_invitations_immutable_guard;

CREATE TRIGGER iam_one_time_tokens_immutable_guard
BEFORE UPDATE ON iam_one_time_tokens
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'purpose', 'user_id', 'identifier_id', 'identifier_kind',
    'identifier_revision', 'invitation_id', 'invitation_expires_at',
    'invitation_revision', 'portal_binding_id', 'portal_setup_revision',
    'selector', 'hash_key_id', 'secret_hmac', 'binding_sha256',
    'created_at', 'expires_at'
);
ALTER TABLE iam_one_time_tokens
    ENABLE ALWAYS TRIGGER iam_one_time_tokens_immutable_guard;

CREATE TRIGGER iam_refresh_families_immutable_guard
BEFORE UPDATE ON iam_refresh_families
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'user_id', 'session_id', 'created_at', 'absolute_expires_at',
    'retain_until'
);
ALTER TABLE iam_refresh_families
    ENABLE ALWAYS TRIGGER iam_refresh_families_immutable_guard;

CREATE TRIGGER iam_refresh_tokens_immutable_guard
BEFORE UPDATE ON iam_refresh_tokens
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'family_id', 'family_absolute_expires_at', 'parent_id',
    'parent_sequence', 'selector', 'hash_key_id', 'secret_hmac', 'sequence',
    'auth_time', 'acr', 'amr_password', 'amr_webauthn', 'amr_totp',
    'amr_recovery', 'step_up_purpose_sha256', 'issued_at', 'idle_expires_at',
    'retain_until'
);
ALTER TABLE iam_refresh_tokens
    ENABLE ALWAYS TRIGGER iam_refresh_tokens_immutable_guard;
