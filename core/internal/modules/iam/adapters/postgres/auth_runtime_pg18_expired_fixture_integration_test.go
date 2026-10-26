//go:build integration

package postgres

const seedIAMExpiredSessionSQL = `
INSERT INTO iam_sessions (
 id,user_id,access_mode_id,access_mode,state,device_kind,browser_family,
 os_family,ip_hmac_key_id,ip_hmac,revision,created_at,updated_at,last_seen_at,
 absolute_expires_at,retain_until
) VALUES (
 '62000000-0000-4000-8000-000000000002',
 '51000000-0000-4000-8000-000000000001',
 '52000000-0000-4000-8000-000000000001','operator_console','active',
 'desktop','chromium','windows','network-v1',
 pg_catalog.decode(pg_catalog.repeat('a2',32),'hex'),1,
 pg_catalog.transaction_timestamp()-INTERVAL '2 hours',
 pg_catalog.transaction_timestamp()-INTERVAL '2 hours',
 pg_catalog.transaction_timestamp()-INTERVAL '2 hours',
	 pg_catalog.transaction_timestamp()-INTERVAL '2 hours'+INTERVAL '90 days',
	 pg_catalog.transaction_timestamp()-INTERVAL '2 hours'+INTERVAL '120 days'
)`

const seedIAMExpiredFamilySQL = `
INSERT INTO iam_refresh_families (
 id,user_id,session_id,state,rotation_sequence,revision,created_at,updated_at,
 absolute_expires_at,retain_until
) VALUES (
 '63000000-0000-4000-8000-000000000002',
 '51000000-0000-4000-8000-000000000001',
 '62000000-0000-4000-8000-000000000002','active',0,1,
 pg_catalog.transaction_timestamp()-INTERVAL '2 hours',
 pg_catalog.transaction_timestamp()-INTERVAL '2 hours',
	 pg_catalog.transaction_timestamp()-INTERVAL '2 hours'+INTERVAL '90 days',
	 pg_catalog.transaction_timestamp()-INTERVAL '2 hours'+INTERVAL '91 days'
)`

const seedIAMExpiredTokenSQL = `
INSERT INTO iam_refresh_tokens (
 id,family_id,family_absolute_expires_at,selector,hash_key_id,secret_hmac,
 sequence,auth_time,acr,amr_password,amr_webauthn,amr_totp,amr_recovery,
 state,revision,issued_at,updated_at,idle_expires_at,retain_until
) SELECT
 '64000000-0000-4000-8000-000000000002',
 family.id,family.absolute_expires_at,
 $1,'refresh-v1',$2,
 1,pg_catalog.transaction_timestamp()-INTERVAL '2 hours','login',true,false,
 false,false,'active',1,pg_catalog.transaction_timestamp()-INTERVAL '2 hours',
 pg_catalog.transaction_timestamp()-INTERVAL '2 hours',
 pg_catalog.transaction_timestamp()-INTERVAL '1 hour',
 family.retain_until
FROM iam_refresh_families AS family
WHERE family.id='63000000-0000-4000-8000-000000000002'`
