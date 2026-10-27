package billingv1

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

type orderAvailability struct {
	events *[]string
	err    error
}

func (policy *orderAvailability) AdmitRoute(context.Context, Route) error {
	*policy.events = append(*policy.events, "availability")
	return policy.err
}

type orderRetirement struct {
	events     *[]string
	admitCalls int
	hitCalls   int
	hitErr     error
}

func (runtime *orderRetirement) AdmitManifest(
	_ context.Context,
	batch contractretirement.Batch,
) (contractretirement.Decision, error) {
	runtime.admitCalls++
	return contractretirement.SealDecision(batch, nil)
}

func (runtime *orderRetirement) RecordHits(
	_ context.Context,
	batch contractretirement.Batch,
) (contractretirement.Decision, error) {
	*runtime.events = append(*runtime.events, "hit")
	runtime.hitCalls++
	if runtime.hitErr != nil {
		return contractretirement.Decision{}, runtime.hitErr
	}
	return contractretirement.SealDecision(batch, nil)
}

type orderPublic struct{ events *[]string }

func (public *orderPublic) ListPublicOfferings(
	context.Context,
	billingapi.PublicQuery,
) (billingapi.QueryPage, error) {
	*public.events = append(*public.events, "operation")
	return billingapi.QueryPage{}, nil
}

func (*orderPublic) GetActiveCampaign(context.Context) (*billingapi.CampaignView, error) {
	return nil, billingapi.ErrNotFound
}

func TestRequestOrderIsAvailabilityThenHitThenOperation(t *testing.T) {
	t.Parallel()
	events := []string{}
	availability := &orderAvailability{events: &events}
	retirement := &orderRetirement{events: &events}
	handler := newOrderHandler(t, availability, retirement, &events)
	if err := handler.Admit(context.Background()); err != nil {
		t.Fatal(err)
	}
	events = events[:0]
	response := servePublicPlans(t, handler)
	if response.Code != 200 || retirement.admitCalls != 1 || retirement.hitCalls != 1 ||
		len(events) != 3 || events[0] != "availability" ||
		events[1] != "hit" || events[2] != "operation" {
		t.Fatalf("response=%d admit=%d hit=%d events=%v",
			response.Code, retirement.admitCalls, retirement.hitCalls, events)
	}
}

func TestAvailabilityAndHitFailuresStopBeforeOperation(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, first, second string
		availabilityErr     error
		hitErr              error
		status              int
	}{
		{"availability", "availability", "", ErrRouteUnavailable, nil, 403},
		{"hit", "availability", "hit", nil, errors.New("evidence failed"), 503},
	} {
		t.Run(test.name, func(t *testing.T) {
			events := []string{}
			availability := &orderAvailability{events: &events, err: test.availabilityErr}
			retirement := &orderRetirement{events: &events, hitErr: test.hitErr}
			handler := newOrderHandler(t, availability, retirement, &events)
			if err := handler.Admit(context.Background()); err != nil {
				t.Fatal(err)
			}
			events = events[:0]
			response := servePublicPlans(t, handler)
			if response.Code != test.status || len(events) < 1 || events[0] != test.first ||
				test.second == "" && len(events) != 1 ||
				test.second != "" && (len(events) != 2 || events[1] != test.second) {
				t.Fatalf("response=%d events=%v", response.Code, events)
			}
		})
	}
}

func newOrderHandler(
	t *testing.T,
	availability AvailabilityPolicy,
	retirement contractretirement.RuntimePort,
	events *[]string,
) *Handler {
	t.Helper()
	manifest, err := RetirementManifest()
	if err != nil {
		t.Fatal(err)
	}
	route := Routes()[0]
	key, err := RetirementKey(route)
	if err != nil {
		t.Fatal(err)
	}
	identity, _ := RetirementIdentity()
	batch, err := contractretirement.NewBatch(identity, []contractretirement.Key{key})
	if err != nil {
		t.Fatal(err)
	}
	return &Handler{
		public: &orderPublic{events: events}, retirement: retirement,
		availability: availability, query: billingapi.NewQueryDecoder(),
		manifest: manifest, keys: map[string]contractretirement.Batch{route.ID: batch},
	}
}

func servePublicPlans(t *testing.T, handler *Handler) *httptest.ResponseRecorder {
	t.Helper()
	bindings, err := handler.BindRoutes()
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/api/billing/plans", nil)
	bindings[0].Handler.ServeHTTP(response, request)
	return response
}
