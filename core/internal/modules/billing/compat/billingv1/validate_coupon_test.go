package billingv1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

type validationIdentityStub struct {
	legacyID, canonicalID string
	err                   error
	calls                 int
}

func (stub *validationIdentityStub) ResolveLegacyIdentity(
	context.Context, billingapi.LegacyIdentityQuery,
) (billingapi.LegacyIdentityView, error) {
	return billingapi.LegacyIdentityView{}, billingapi.ErrUnavailable
}

func (stub *validationIdentityStub) ResolvePublicLegacyIdentity(
	_ context.Context, kind billingapi.LegacyIdentityKind, identifier string,
) (billingapi.LegacyIdentityView, error) {
	stub.calls++
	if stub.err != nil || kind != billingapi.LegacyIdentityOffering ||
		identifier != stub.legacyID {
		return billingapi.LegacyIdentityView{}, stub.err
	}
	return billingapi.LegacyIdentityView{
		LegacyID: stub.legacyID, CanonicalID: stub.canonicalID,
	}, nil
}

type validationPublicStub struct {
	offering *billingapi.OfferingView
	calls    int
}

func (stub *validationPublicStub) ListPublicOfferings(
	context.Context, billingapi.PublicQuery,
) (billingapi.QueryPage, error) {
	stub.calls++
	page := billingapi.QueryPage{}
	if stub.offering != nil {
		document, _ := json.Marshal(stub.offering)
		page.Items = []json.RawMessage{document}
	}
	return page, nil
}

func (*validationPublicStub) GetActiveCampaign(
	context.Context,
) (*billingapi.CampaignView, error) {
	return nil, billingapi.ErrNotFound
}

func TestValidationOfferingIDResolvesAdoptedLegacyIdentity(t *testing.T) {
	t.Parallel()
	legacyID, canonicalID := uuid.NewString(), uuid.NewString()
	identity := &validationIdentityStub{legacyID: legacyID, canonicalID: canonicalID}
	public := &validationPublicStub{offering: &billingapi.OfferingView{
		ID: canonicalID, VersionID: uuid.NewString(),
	}}
	handler := &Handler{
		identity: identity, public: public, query: billingapi.NewQueryDecoder(),
	}
	result, err := handler.validationOfferingID(
		httptest.NewRequest("POST", "/", nil), legacyID,
	)
	if err != nil || result != public.offering.VersionID {
		t.Fatalf("offering version = %q, error = %v", result, err)
	}
}

func TestValidationOfferingIDSkipsAbsentAndZeroPlan(t *testing.T) {
	t.Parallel()
	identity := &validationIdentityStub{}
	public := &validationPublicStub{}
	handler := &Handler{identity: identity, public: public}
	for _, planID := range []string{"", uuid.Nil.String()} {
		result, err := handler.validationOfferingID(
			httptest.NewRequest("POST", "/", nil), planID,
		)
		if err != nil || result != "" {
			t.Fatalf("plan %q resolved as %q: %v", planID, result, err)
		}
	}
	if identity.calls != 0 || public.calls != 0 {
		t.Fatalf("optional plan consulted identity=%d public=%d", identity.calls, public.calls)
	}
}

func TestValidationOfferingIDFailsClosedForWrongKindOrMissingPlan(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"wrong-kind", "not-found"} {
		t.Run(name, func(t *testing.T) {
			identity := &validationIdentityStub{
				legacyID: uuid.NewString(), err: billingapi.ErrNotFound,
			}
			handler := &Handler{identity: identity, public: &validationPublicStub{}}
			_, err := handler.validationOfferingID(
				httptest.NewRequest("POST", "/", nil), identity.legacyID,
			)
			if !errors.Is(err, billingapi.ErrNotFound) {
				t.Fatalf("%s error = %v", name, err)
			}
		})
	}
}
