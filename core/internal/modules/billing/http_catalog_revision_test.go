package billing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

const offeringVersionDocument = `{
  "expected_revision":11,
  "name":"Pro",
  "public":true,
  "base":false,
  "trial_days":14,
  "grace_days":3,
  "effective_start":"2026-08-01T00:00:00.000000Z",
  "prices":[{"cadence":"monthly","minor_units":9900,"currency":"TRY","currency_registry_version":1}],
  "entitlements":[]
}`

func TestCreateOfferingVersionCarriesExactExpectedRevision(t *testing.T) {
	var captured billingapi.CreateOfferingVersionCommand
	handler := newHandlerFixture(t, nil, catalogStub{createOfferingVersion: func(
		_ context.Context,
		command billingapi.CreateOfferingVersionCommand,
	) (billingapi.OfferingView, error) {
		captured = command
		return billingapi.OfferingView{ID: command.OfferingID, Revision: 12}, nil
	}})
	response := serveOfferingVersion(handler, offeringVersionDocument)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	if captured.ExpectedRevision != 11 ||
		captured.Version.EffectiveStart != "2026-08-01T00:00:00.000000Z" {
		t.Fatalf("command mismatch: %#v", captured)
	}
}

func TestCreateOfferingVersionRejectsMissingRevisionBeforePort(t *testing.T) {
	called := false
	handler := newHandlerFixture(t, nil, catalogStub{createOfferingVersion: func(
		context.Context,
		billingapi.CreateOfferingVersionCommand,
	) (billingapi.OfferingView, error) {
		called = true
		return billingapi.OfferingView{}, nil
	}})
	document := strings.Replace(offeringVersionDocument, `"expected_revision":11,`, "", 1)
	response := serveOfferingVersion(handler, document)
	if response.Code != http.StatusBadRequest || called {
		t.Fatalf("status=%d called=%t body=%s", response.Code, called, response.Body.String())
	}
}

func TestCreateOfferingVersionMapsStaleRevisionToConflict(t *testing.T) {
	handler := newHandlerFixture(t, nil, catalogStub{createOfferingVersion: func(
		context.Context,
		billingapi.CreateOfferingVersionCommand,
	) (billingapi.OfferingView, error) {
		return billingapi.OfferingView{}, billingapi.ErrRevisionStale
	}})
	response := serveOfferingVersion(handler, offeringVersionDocument)
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusConflict || envelope.Error.Code != "BILLING_REVISION_STALE" {
		t.Fatalf("status=%d envelope=%#v", response.Code, envelope)
	}
}

func serveOfferingVersion(handler *Handler, document string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/platform/billing/offerings/00000000-0000-4000-8000-000000000010/versions",
		strings.NewReader(document),
	)
	request.SetPathValue("offering_id", "00000000-0000-4000-8000-000000000010")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "billing-offering-version-11")
	response := httptest.NewRecorder()
	handler.createOfferingVersion(response, request)
	return response
}
