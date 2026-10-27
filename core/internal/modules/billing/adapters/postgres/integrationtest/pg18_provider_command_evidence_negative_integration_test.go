//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertProviderPGWildcardClaimRejected(
	t *testing.T,
	ctx context.Context,
	f providerPGFixture,
	attemptID uuid.UUID,
) {
	t.Helper()
	err := f.connection.QueryRowContext(ctx, `
SELECT job_state FROM platform.platform_async_external_effect_evidence_v1(
 $1,'tenant',$1,'billing_forms',1,'billing',1,$2,$3,1,
 'billing.provider.finalize',1,$4,$5,'billing:provider:%',$6,
 'billing.provider.command',1,$7,$5,$8,$8,'testpay','command',$6
)`, f.tenantID, f.finalizerID, attemptID,
		"billing-provider-finalize/"+f.commandID.String(), f.digestHex, f.commandID,
		"billing-provider-command/testpay/"+f.commandID.String(),
		"billing:provider:command:testpay:"+f.commandID.String(),
	).Scan(new(string))
	assertProviderPGCode(t, err, "P5721")
}

func assertProviderPGMissingAttemptRejected(
	t *testing.T,
	ctx context.Context,
	f providerPGFixture,
	attemptID uuid.UUID,
) {
	t.Helper()
	transaction, err := f.connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `
UPDATE platform.jobs SET attempts=1,lease_generation=1 WHERE id=$1
`, f.commandID); err != nil {
		t.Fatal(err)
	}
	_, _, _, err = submitProviderPGEvidence(ctx, transaction, f, attemptID)
	assertProviderPGCode(t, err, "P5721")
}

func assertProviderPGTerminalZeroAttemptRejected(
	t *testing.T,
	ctx context.Context,
	f providerPGFixture,
	attemptID uuid.UUID,
) {
	t.Helper()
	transaction, err := f.connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	completed := f.now.Add(2 * time.Second)
	if _, err = transaction.ExecContext(ctx, `
UPDATE platform.jobs SET state='succeeded',updated_at=$2::timestamptz,
 completed_at=$2::timestamptz,retain_until=$2::timestamptz+interval '91 days'
WHERE id=$1
`, f.commandID, completed); err != nil {
		t.Fatal(err)
	}
	_, _, _, err = submitProviderPGEvidence(ctx, transaction, f, attemptID)
	assertProviderPGCode(t, err, "P5721")
}

func submitProviderPGEvidence(
	ctx context.Context,
	runner providerPGRunner,
	f providerPGFixture,
	attemptID uuid.UUID,
) (string, int, bool, error) {
	var state string
	var attempts int
	var started bool
	err := runner.QueryRowContext(ctx, `
SELECT job_state,job_attempts,effect_started
FROM platform.platform_async_external_effect_evidence_v1(
 $1,'tenant',$1,'billing_forms',1,'billing',1,$2,$3,1,
 'billing.provider.finalize',1,$4,$5,$6,$7,'billing.provider.command',1,
 $8,$5,$9,$9,'testpay','command',$7
)`, f.tenantID, f.finalizerID, attemptID,
		"billing-provider-finalize/"+f.commandID.String(), f.digestHex,
		"billing:provider:finalize:"+f.commandID.String(), f.commandID,
		"billing-provider-command/testpay/"+f.commandID.String(),
		"billing:provider:command:testpay:"+f.commandID.String(),
	).Scan(&state, &attempts, &started)
	return state, attempts, started, err
}
