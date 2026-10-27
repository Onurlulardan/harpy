package manual

import (
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
)

const (
	manualTenantID  = "11111111-1111-4111-8111-111111111111"
	manualEffectID  = "22222222-2222-4222-8222-222222222222"
	manualCausation = "33333333-3333-4333-8333-333333333333"
	manualDigest    = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

var manualNow = time.Date(2026, 7, 17, 12, 0, 0, 321000000, time.UTC)

type manualClock struct{ value time.Time }

func (clock manualClock) Now() time.Time { return clock.value }

func manualRequest(t *testing.T, operation string) effectsv1.Request {
	t.Helper()
	tenantID := manualTenantID
	request, err := effectsv1.NewRequest(
		manualEffectID, "billing.provider.command", 1, &tenantID,
		effectsv1.Authority{
			ScopeKind: effectsv1.ScopeTenant, ScopeID: tenantID,
			GroupID: "billing", GroupGeneration: 2,
			Domain: "billing", CutoverGeneration: 3,
			WriterOrigin: effectsv1.WriterV2,
		}, "manual", operation, "manual:subscription:10", "corr-10",
		manualCausation, "billing-provider-command/10", manualDigest,
		"2026-07-17T11:59:59Z",
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}
