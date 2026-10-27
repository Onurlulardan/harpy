package billing

import (
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/provider/manual"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
)

const (
	providerTenantID = "11111111-1111-4111-8111-111111111111"
	providerEffectID = "22222222-2222-4222-8222-222222222222"
	providerCauseID  = "33333333-3333-4333-8333-333333333333"
	providerDigest   = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

type providerClock struct{ value time.Time }

func (clock providerClock) Now() time.Time { return clock.value }

func providerRegistry(t *testing.T) providerapp.Registry {
	t.Helper()
	adapter, err := manual.New(providerClock{time.Date(
		2026, 7, 17, 12, 0, 0, 0, time.UTC,
	)})
	if err != nil {
		t.Fatal(err)
	}
	registry, err := providerapp.NewRegistry(adapter)
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func providerEffectRequest(t *testing.T, operation string) effectsv1.Request {
	t.Helper()
	tenantID := providerTenantID
	request, err := effectsv1.NewRequest(
		providerEffectID, "billing.provider.command", 1, &tenantID,
		effectsv1.Authority{
			ScopeKind: effectsv1.ScopeTenant, ScopeID: tenantID,
			GroupID: "billing", GroupGeneration: 1,
			Domain: "billing", CutoverGeneration: 1,
			WriterOrigin: effectsv1.WriterV2,
		}, "manual", operation, "manual:command:1", "corr-1", providerCauseID,
		"billing-provider-command/1", providerDigest, "2026-07-17T11:59:59Z",
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}
