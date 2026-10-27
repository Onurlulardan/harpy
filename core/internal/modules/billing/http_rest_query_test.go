package billing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestOfferingListUsesCanonicalRESTQueryGETAndPOSTDecoders(t *testing.T) {
	var captured []billingapi.Query
	catalog := catalogStub{
		listOfferings: func(
			_ context.Context,
			query billingapi.Query,
		) (billingapi.QueryPage, error) {
			captured = append(captured, query)
			return billingapi.QueryPage{}, nil
		},
		createOfferingVersion: func(
			context.Context,
			billingapi.CreateOfferingVersionCommand,
		) (billingapi.OfferingView, error) {
			return billingapi.OfferingView{}, nil
		},
	}
	handler := newHandlerFixture(t, nil, catalog)
	get := httptest.NewRequest(
		http.MethodGet,
		"/api/platform/billing/offerings?page[size]=25&sort[0][field]=created_at&sort[0][direction]=desc",
		nil,
	)
	getResponse := httptest.NewRecorder()
	handler.listOfferings(getResponse, get)
	post := httptest.NewRequest(
		http.MethodPost,
		"/api/platform/billing/offerings/query",
		strings.NewReader(`{"page":{"size":25,"after":null},"sort":[{"field":"created_at","direction":"desc"}]}`),
	)
	post.Header.Set("Content-Type", "application/json")
	postResponse := httptest.NewRecorder()
	handler.queryOfferings(postResponse, post)
	if getResponse.Code != http.StatusOK || postResponse.Code != http.StatusOK ||
		len(captured) != 2 {
		t.Fatalf("GET=%d POST=%d queries=%d", getResponse.Code, postResponse.Code, len(captured))
	}
	for index, query := range captured {
		if query.Request.Page().Size() != 25 || len(query.Request.Sort()) != 1 ||
			query.Request.Sort()[0].Field() != "created_at" {
			t.Fatalf("query %d mismatch: %#v", index, query.Request)
		}
		if !query.Access.ValidFor(billingapi.PermissionPlatformBillingRead) {
			t.Fatalf("query %d lacks verified platform access", index)
		}
	}
}

func TestOfferingListRejectsMalformedRESTQueryBeforePort(t *testing.T) {
	called := false
	catalog := catalogStub{
		listOfferings: func(
			context.Context,
			billingapi.Query,
		) (billingapi.QueryPage, error) {
			called = true
			return billingapi.QueryPage{}, nil
		},
		createOfferingVersion: func(
			context.Context,
			billingapi.CreateOfferingVersionCommand,
		) (billingapi.OfferingView, error) {
			return billingapi.OfferingView{}, nil
		},
	}
	handler := newHandlerFixture(t, nil, catalog)
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/platform/billing/offerings?page[size]=25&page[size]=50",
		nil,
	)
	response := httptest.NewRecorder()
	handler.listOfferings(response, request)
	if response.Code != http.StatusUnprocessableEntity || called {
		t.Fatalf("status=%d called=%t body=%s", response.Code, called, response.Body.String())
	}
}
