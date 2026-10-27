package billing

import (
	"context"
	"errors"
	"testing"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

type manualWebhookClaim struct {
	tenantID  uuid.UUID
	reference string
}

func (claim manualWebhookClaim) JobID() uuid.UUID     { return uuid.New() }
func (claim manualWebhookClaim) AttemptID() uuid.UUID { return uuid.New() }
func (claim manualWebhookClaim) TenantID() (uuid.UUID, bool) {
	return claim.tenantID, claim.tenantID != uuid.Nil
}
func (manualWebhookClaim) Kind() string                   { return providerapp.WebhookProcessorJobKind }
func (manualWebhookClaim) PayloadRevision() int           { return 1 }
func (claim manualWebhookClaim) PayloadReference() string { return claim.reference }
func (manualWebhookClaim) PayloadSHA256() string          { return providerDigest }
func (manualWebhookClaim) LeaseGeneration() int64         { return 1 }

func TestManualOnlyWebhookJobFailsClosedWithoutCompletingWork(t *testing.T) {
	claim := manualWebhookClaim{
		tenantID:  uuid.MustParse("018f1000-0000-7000-8000-000000000001"),
		reference: "billing-webhook-receipt/018f1000-0000-7000-8000-000000000002",
	}
	outcome, err := (manualOnlyWebhookJob{}).Execute(context.Background(), claim)
	if !errors.Is(err, providerapp.ErrCapability) {
		t.Fatalf("manual-only webhook job error = %v", err)
	}
	if outcome.Valid() {
		t.Fatal("manual-only webhook job accidentally completed")
	}
}
