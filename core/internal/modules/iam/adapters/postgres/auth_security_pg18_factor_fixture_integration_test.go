//go:build integration

package postgres

import (
	"bytes"
	"context"
	"testing"
)

const (
	iamSecurityRecoveryAuthenticatorID = "86000000-0000-4000-8000-000000000001"
	iamSecurityRecoveryCodeID          = "86100000-0000-4000-8000-000000000001"
)

type iamAuthSecurityFactor struct {
	tenant   iamFoundationTenant
	codeHMAC []byte
}

func seedIAMAuthSecurityFactors(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) []iamAuthSecurityFactor {
	t.Helper()
	factors := make([]iamAuthSecurityFactor, 0, len(iamFoundationTenants))
	for index, tenant := range iamFoundationTenants {
		factor := iamAuthSecurityFactor{
			tenant:   tenant,
			codeHMAC: bytes.Repeat([]byte{byte(0xb1 + index)}, 32),
		}
		mustExecIAMTenant(
			t, ctx, database.migrator, tenant.schema,
			seedIAMAuthSecurityRecoveryAuthenticatorSQL,
		)
		mustExecIAMTenant(
			t, ctx, database.migrator, tenant.schema,
			seedIAMAuthSecurityRecoveryCodeSQL, factor.codeHMAC,
		)
		factors = append(factors, factor)
	}
	return factors
}

const seedIAMAuthSecurityRecoveryAuthenticatorSQL = `
INSERT INTO iam_authenticators (
 id,user_id,kind,display_label,state,revision,
 created_at,updated_at,activated_at
) VALUES (
 '86000000-0000-4000-8000-000000000001',
 '51000000-0000-4000-8000-000000000001','recovery_set_v1',
 'Security recovery set','active',1,pg_catalog.transaction_timestamp(),
 pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp()
)`

const seedIAMAuthSecurityRecoveryCodeSQL = `
INSERT INTO iam_recovery_codes (
 id,authenticator_id,user_id,authenticator_kind,position,hash_profile,
 hmac_key_id,code_hmac,state,revision,created_at,updated_at
) VALUES (
 '86100000-0000-4000-8000-000000000001',
 '86000000-0000-4000-8000-000000000001',
 '51000000-0000-4000-8000-000000000001','recovery_set_v1',1,
 'hmac_sha256_v1','iam-recovery-key-v1',$1,'active',1,
 pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp()
)`
