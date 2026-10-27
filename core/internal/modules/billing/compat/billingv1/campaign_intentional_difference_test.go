package billingv1

import (
	"errors"
	"testing"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestCampaignSlugMutationIsStableIntentionalDifference(t *testing.T) {
	replacement := "renamed-summer"
	_, err := mapCampaignUpdate(
		billingapi.Access{}, "key",
		billingapi.CampaignView{Key: "summer"},
		campaignRequest{Slug: &replacement}, time.Now(),
	)
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("slug mutation error=%v", err)
	}
	status, message, code := failureDetails(
		"billing.route.admin-campaign-update", err,
	)
	if status != 422 ||
		message != "Mutation is not supported by the immutable Billing model" ||
		code != "BILLING_MUTATION_UNSUPPORTED" {
		t.Fatalf("status=%d message=%q code=%q", status, message, code)
	}
}
