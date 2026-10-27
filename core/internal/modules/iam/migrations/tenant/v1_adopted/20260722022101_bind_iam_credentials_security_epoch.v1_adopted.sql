-- +goose Up
UPDATE iam_sessions SET
    state = 'revoked', revision = revision + 1,
    updated_at = pg_catalog.transaction_timestamp(),
    revoked_at = pg_catalog.transaction_timestamp(),
    retain_until = GREATEST(
        retain_until, pg_catalog.transaction_timestamp() + INTERVAL '30 days'
    )
WHERE state = 'active';

UPDATE iam_refresh_families SET
    state = 'revoked', revoke_reason = 'security_epoch_cutover',
    revision = revision + 1, updated_at = pg_catalog.transaction_timestamp(),
    revoked_at = pg_catalog.transaction_timestamp()
WHERE state = 'active';

UPDATE iam_refresh_tokens SET
    state = 'revoked', revision = revision + 1,
    updated_at = pg_catalog.transaction_timestamp(),
    revoked_at = pg_catalog.transaction_timestamp()
WHERE state = 'active';

UPDATE iam_one_time_tokens SET
    state = 'revoked', revision = revision + 1,
    updated_at = pg_catalog.transaction_timestamp(),
    revoked_at = pg_catalog.transaction_timestamp(),
    purge_after = GREATEST(
        purge_after, pg_catalog.transaction_timestamp() + INTERVAL '24 hours'
    )
WHERE state = 'pending';

UPDATE iam_auth_challenges SET
    state = 'cancelled', revision = revision + 1,
    updated_at = pg_catalog.transaction_timestamp(),
    terminal_at = pg_catalog.transaction_timestamp(),
    purge_after = GREATEST(
        purge_after, pg_catalog.transaction_timestamp() + INTERVAL '24 hours'
    )
WHERE state = 'pending';

ALTER TABLE iam_sessions ADD COLUMN security_epoch bytea;
ALTER TABLE iam_refresh_families ADD COLUMN security_epoch bytea;
ALTER TABLE iam_refresh_tokens ADD COLUMN security_epoch bytea;
ALTER TABLE iam_one_time_tokens ADD COLUMN security_epoch bytea;
ALTER TABLE iam_auth_challenges ADD COLUMN security_epoch bytea;

UPDATE iam_sessions
SET security_epoch=iam_security_epoch_current_v1(pg_catalog.current_schema());
UPDATE iam_refresh_families
SET security_epoch=iam_security_epoch_current_v1(pg_catalog.current_schema());
UPDATE iam_refresh_tokens
SET security_epoch=iam_security_epoch_current_v1(pg_catalog.current_schema());
UPDATE iam_one_time_tokens
SET security_epoch=iam_security_epoch_current_v1(pg_catalog.current_schema());
UPDATE iam_auth_challenges
SET security_epoch=iam_security_epoch_current_v1(pg_catalog.current_schema());

ALTER TABLE iam_sessions ALTER COLUMN security_epoch SET NOT NULL,
    ALTER COLUMN security_epoch SET DEFAULT
      iam_security_epoch_current_v1(pg_catalog.current_schema());
ALTER TABLE iam_refresh_families ALTER COLUMN security_epoch SET NOT NULL,
    ALTER COLUMN security_epoch SET DEFAULT
      iam_security_epoch_current_v1(pg_catalog.current_schema());
ALTER TABLE iam_refresh_tokens ALTER COLUMN security_epoch SET NOT NULL,
    ALTER COLUMN security_epoch SET DEFAULT
      iam_security_epoch_current_v1(pg_catalog.current_schema());
ALTER TABLE iam_one_time_tokens ALTER COLUMN security_epoch SET NOT NULL,
    ALTER COLUMN security_epoch SET DEFAULT
      iam_security_epoch_current_v1(pg_catalog.current_schema());
ALTER TABLE iam_auth_challenges ALTER COLUMN security_epoch SET NOT NULL,
    ALTER COLUMN security_epoch SET DEFAULT
      iam_security_epoch_current_v1(pg_catalog.current_schema());

ALTER TABLE iam_sessions ADD CONSTRAINT iam_sessions_security_epoch_ck
    CHECK (pg_catalog.octet_length(security_epoch) = 32);
ALTER TABLE iam_refresh_families ADD CONSTRAINT iam_refresh_families_security_epoch_ck
    CHECK (pg_catalog.octet_length(security_epoch) = 32);
ALTER TABLE iam_refresh_tokens ADD CONSTRAINT iam_refresh_tokens_security_epoch_ck
    CHECK (pg_catalog.octet_length(security_epoch) = 32);
ALTER TABLE iam_one_time_tokens ADD CONSTRAINT iam_one_time_tokens_security_epoch_ck
    CHECK (pg_catalog.octet_length(security_epoch) = 32);
ALTER TABLE iam_auth_challenges ADD CONSTRAINT iam_auth_challenges_security_epoch_ck
    CHECK (pg_catalog.octet_length(security_epoch) = 32);
