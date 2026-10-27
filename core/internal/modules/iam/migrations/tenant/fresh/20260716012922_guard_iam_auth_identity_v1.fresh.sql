-- +goose Up
CREATE TRIGGER iam_sessions_immutable_guard
BEFORE UPDATE ON iam_sessions
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'user_id', 'access_mode_id', 'access_mode', 'portal_binding_id',
    'device_kind', 'browser_family', 'os_family', 'ip_hmac_key_id', 'ip_hmac',
    'created_at', 'absolute_expires_at', 'retain_until'
);
ALTER TABLE iam_sessions ENABLE ALWAYS TRIGGER iam_sessions_immutable_guard;

CREATE TRIGGER iam_authenticators_immutable_guard
BEFORE UPDATE ON iam_authenticators
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'user_id', 'kind', 'created_at'
);
ALTER TABLE iam_authenticators
    ENABLE ALWAYS TRIGGER iam_authenticators_immutable_guard;

CREATE TRIGGER iam_webauthn_credentials_immutable_guard
BEFORE UPDATE ON iam_webauthn_credentials
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'authenticator_id', 'user_id', 'authenticator_kind', 'credential_id',
    'public_key_cose', 'rp_id', 'rp_id_sha256', 'origin_policy_version',
    'aaguid', 'backup_eligible', 'discoverable', 'uv_required', 'created_at'
);
ALTER TABLE iam_webauthn_credentials
    ENABLE ALWAYS TRIGGER iam_webauthn_credentials_immutable_guard;

CREATE TRIGGER iam_totp_authenticators_immutable_guard
BEFORE UPDATE ON iam_totp_authenticators
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'authenticator_id', 'user_id', 'authenticator_kind', 'profile',
    'cipher_suite', 'kms_key_id', 'nonce', 'secret_ciphertext',
    'authentication_tag', 'aad_sha256', 'enrollment_expires_at', 'created_at'
);
ALTER TABLE iam_totp_authenticators
    ENABLE ALWAYS TRIGGER iam_totp_authenticators_immutable_guard;

CREATE TRIGGER iam_recovery_codes_immutable_guard
BEFORE UPDATE ON iam_recovery_codes
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'authenticator_id', 'user_id', 'authenticator_kind', 'position',
    'hash_profile', 'hmac_key_id', 'code_hmac', 'created_at'
);
ALTER TABLE iam_recovery_codes
    ENABLE ALWAYS TRIGGER iam_recovery_codes_immutable_guard;
