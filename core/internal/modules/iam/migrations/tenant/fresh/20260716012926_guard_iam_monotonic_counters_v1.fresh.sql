-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_monotonic_bigint_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    column_name text;
    old_value bigint;
    new_value bigint;
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR TG_NARGS <> 1 THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'IAM monotonic guard invocation is invalid';
    END IF;
    column_name := TG_ARGV[0];
    old_value := (pg_catalog.to_jsonb(OLD) ->> column_name)::bigint;
    new_value := (pg_catalog.to_jsonb(NEW) ->> column_name)::bigint;
    IF old_value IS NOT NULL AND
       (new_value IS NULL OR new_value < old_value) THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            CONSTRAINT = TG_TABLE_NAME || '_' || column_name ||
                '_monotonic_guard',
            MESSAGE = pg_catalog.format(
                'IAM monotonic column %I cannot decrease',
                column_name
            );
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_monotonic_bigint_v1() FROM PUBLIC;

CREATE TRIGGER iam_password_credentials_version_guard
BEFORE UPDATE ON iam_password_credentials
FOR EACH ROW EXECUTE FUNCTION iam_guard_monotonic_bigint_v1(
    'credential_version'
);
ALTER TABLE iam_password_credentials
    ENABLE ALWAYS TRIGGER iam_password_credentials_version_guard;

CREATE TRIGGER iam_webauthn_credentials_sign_count_guard
BEFORE UPDATE ON iam_webauthn_credentials
FOR EACH ROW EXECUTE FUNCTION iam_guard_monotonic_bigint_v1(
    'sign_count'
);
ALTER TABLE iam_webauthn_credentials
    ENABLE ALWAYS TRIGGER iam_webauthn_credentials_sign_count_guard;

CREATE TRIGGER iam_totp_authenticators_last_step_guard
BEFORE UPDATE ON iam_totp_authenticators
FOR EACH ROW EXECUTE FUNCTION iam_guard_monotonic_bigint_v1(
    'last_accepted_time_step'
);
ALTER TABLE iam_totp_authenticators
    ENABLE ALWAYS TRIGGER iam_totp_authenticators_last_step_guard;

CREATE TRIGGER iam_auth_challenges_attempt_guard
BEFORE UPDATE ON iam_auth_challenges
FOR EACH ROW EXECUTE FUNCTION iam_guard_monotonic_bigint_v1(
    'attempt_count'
);
ALTER TABLE iam_auth_challenges
    ENABLE ALWAYS TRIGGER iam_auth_challenges_attempt_guard;

CREATE TRIGGER iam_one_time_tokens_attempt_guard
BEFORE UPDATE ON iam_one_time_tokens
FOR EACH ROW EXECUTE FUNCTION iam_guard_monotonic_bigint_v1(
    'attempt_count'
);
ALTER TABLE iam_one_time_tokens
    ENABLE ALWAYS TRIGGER iam_one_time_tokens_attempt_guard;

CREATE TRIGGER iam_refresh_families_rotation_guard
BEFORE UPDATE ON iam_refresh_families
FOR EACH ROW EXECUTE FUNCTION iam_guard_monotonic_bigint_v1(
    'rotation_sequence'
);
ALTER TABLE iam_refresh_families
    ENABLE ALWAYS TRIGGER iam_refresh_families_rotation_guard;
