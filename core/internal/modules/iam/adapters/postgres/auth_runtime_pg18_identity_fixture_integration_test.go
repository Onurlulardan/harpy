//go:build integration

package postgres

const seedIAMAuthUserStateSQL = `
UPDATE iam_users SET authorization_version=1`

const seedIAMAuthIdentifierSQL = `
INSERT INTO iam_user_identifiers (
 id,user_id,kind,display_value,normalized_value,normalized_sha256,state,
 revision,verified_at,activated_at,created_at,updated_at
) VALUES (
 '61000000-0000-4000-8000-000000000001',
 '51000000-0000-4000-8000-000000000001','email_v1',
 'Runtime@example.com','runtime@example.com',$1,'active',1,
 pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp(),
 pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp()
)`

const seedIAMAuthPasswordSQL = `
INSERT INTO iam_password_credentials (
 user_id,credential_version,state,algorithm,salt,tag,memory_kib,iterations,
 parallelism,legacy_bcrypt_hash,upgrade_required,revision,created_at,updated_at
) VALUES (
 '51000000-0000-4000-8000-000000000001',1,'active','argon2id_v1',
 pg_catalog.decode(pg_catalog.repeat('a1',16),'hex'),$1,65536,3,4,NULL,
 false,1,pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp()
)`
