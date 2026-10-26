package billing

import (
	"context"
	"testing"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestPlatformOperationsRejectMissingExactAccessEvidence(t *testing.T) {
	runtime := &GlobalHTTPRuntime{}
	ctx := context.Background()
	identifier := "019c92a0-0000-7000-8000-000000000099"
	checks := []func() error{
		func() error { _, err := runtime.ListOfferings(ctx, billingapi.Query{}); return err },
		func() error { _, err := runtime.GetOffering(ctx, billingapi.Access{}, identifier); return err },
		func() error { _, err := runtime.CreateOffering(ctx, billingapi.CreateOfferingCommand{}); return err },
		func() error {
			_, err := runtime.CreateOfferingVersion(ctx, billingapi.CreateOfferingVersionCommand{})
			return err
		},
		func() error { _, err := runtime.RetireOffering(ctx, billingapi.RetireOfferingCommand{}); return err },
		func() error { _, err := runtime.ListDiscounts(ctx, billingapi.Query{}); return err },
		func() error { _, err := runtime.GetDiscount(ctx, billingapi.Access{}, identifier); return err },
		func() error { _, err := runtime.CreateDiscount(ctx, billingapi.CreateDiscountCommand{}); return err },
		func() error {
			_, err := runtime.CreateDiscountVersion(ctx, billingapi.CreateDiscountVersionCommand{})
			return err
		},
		func() error { _, err := runtime.RetireDiscount(ctx, billingapi.RetireDiscountCommand{}); return err },
		func() error { _, err := runtime.ListCampaigns(ctx, billingapi.Query{}); return err },
		func() error { _, err := runtime.GetCampaign(ctx, billingapi.Access{}, identifier); return err },
		func() error { _, err := runtime.CreateCampaign(ctx, billingapi.CreateCampaignCommand{}); return err },
		func() error {
			_, err := runtime.CreateCampaignVersion(ctx, billingapi.CreateCampaignVersionCommand{})
			return err
		},
		func() error { _, err := runtime.RetireCampaign(ctx, billingapi.RetireCampaignCommand{}); return err },
	}
	for index, check := range checks {
		if err := check(); err == nil {
			t.Fatalf("operation %d accepted missing access", index)
		}
	}
}

func TestPublicOperationsRejectInvalidRequestOrClockBeforeStorage(t *testing.T) {
	runtime := &GlobalHTTPRuntime{clock: func() time.Time { return time.Time{} }}
	if _, err := runtime.ListPublicOfferings(context.Background(), billingapi.PublicQuery{}); err == nil {
		t.Fatal("invalid public query was accepted")
	}
	if _, err := runtime.GetActiveCampaign(context.Background()); err == nil {
		t.Fatal("invalid public clock was accepted")
	}
}
