package billingv1

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type bodyDecoder func(*http.Request) error

func TestBodyRoutesRejectUnknownTopLevelMembers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		decode bodyDecoder
	}{
		{"change-plan", decodeChangePlanBody},
		{"cancel", decodeCancelBody},
		{"resume", decodeResumeBody},
		{"apply-coupon", decodeApplyCouponBody},
		{"validate-coupon", decodeValidateCouponBody},
		{"plan-create", decodePlanBody},
		{"plan-update", decodePlanBody},
		{"subscription-update", decodeSubscriptionUpdateBody},
		{"coupon-create", decodeCouponBody},
		{"coupon-update", decodeCouponBody},
		{"campaign-create", decodeCampaignBody},
		{"campaign-update", decodeCampaignBody},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest(
				http.MethodPost, "/", strings.NewReader(`{"unknown":true}`),
			)
			if err := test.decode(request); !errors.Is(err, ErrRequestInvalid) {
				t.Fatalf("unknown member error = %v", err)
			}
		})
	}
}

func TestPlanBodyKeepsMapLikeLimitAndFeatureObjects(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"limits":{"custom_limit":12},"features":{"custom_flag":true}}`,
	))
	if err := decodePlanBody(request); err != nil {
		t.Fatalf("map-like plan members rejected: %v", err)
	}
}

func decodeChangePlanBody(r *http.Request) error {
	_, err := decodeBody[changePlanRequest](r)
	return err
}
func decodeCancelBody(r *http.Request) error {
	_, err := decodeBody[cancelRequest](r)
	return err
}
func decodeResumeBody(r *http.Request) error { return decodeOptionalBody[struct{}](r) }
func decodeApplyCouponBody(r *http.Request) error {
	_, err := decodeBody[applyCouponRequest](r)
	return err
}
func decodeValidateCouponBody(r *http.Request) error {
	_, err := decodeBody[validateCouponRequest](r)
	return err
}
func decodePlanBody(r *http.Request) error {
	_, err := decodeBody[planRequest](r)
	return err
}
func decodeSubscriptionUpdateBody(r *http.Request) error {
	_, err := decodeBody[subscriptionUpdateRequest](r)
	return err
}
func decodeCouponBody(r *http.Request) error {
	_, err := decodeBody[couponRequest](r)
	return err
}
func decodeCampaignBody(r *http.Request) error {
	_, err := decodeBody[campaignRequest](r)
	return err
}
