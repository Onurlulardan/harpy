package billing

import (
	"context"
	"fmt"
	"strings"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type manualOnlyWebhookJob struct{}

func newManualOnlyWebhookJobBinding() (platformasync.DatabaseJobBinding, error) {
	return platformasync.NewDatabaseJobBinding(
		"billing",
		func(platformtx.JobBoundExecutor) platformasync.DatabaseJobCommand {
			return manualOnlyWebhookJob{}
		},
	)
}

func (manualOnlyWebhookJob) Execute(
	_ context.Context,
	claim platformasync.DatabaseJobClaim,
) (platformasync.DatabaseJobOutcome, error) {
	if !validWebhookJobClaim(claim) {
		return platformasync.DatabaseJobOutcome{},
			fmt.Errorf("manual-only billing webhook job claim is invalid")
	}
	return platformasync.DatabaseJobOutcome{}, providerapp.ErrCapability
}

func validWebhookJobClaim(claim platformasync.DatabaseJobClaim) bool {
	if claim == nil || claim.Kind() != providerapp.WebhookProcessorJobKind ||
		claim.PayloadRevision() != 1 ||
		!strings.HasPrefix(claim.PayloadReference(), "billing-webhook-receipt/") {
		return false
	}
	tenantID, tenantScoped := claim.TenantID()
	receiptID, err := uuid.Parse(strings.TrimPrefix(
		claim.PayloadReference(), "billing-webhook-receipt/",
	))
	return err == nil && tenantScoped && tenantID != uuid.Nil && receiptID != uuid.Nil
}
