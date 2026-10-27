package httpedge

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/requestattribution"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type observationOrderWriter struct {
	header    http.Header
	status    int
	committed bool
}

func (writer *observationOrderWriter) Header() http.Header { return writer.header }

func (writer *observationOrderWriter) WriteHeader(status int) {
	writer.status, writer.committed = status, true
}

func (writer *observationOrderWriter) Write(body []byte) (int, error) {
	writer.committed = true
	return len(body), nil
}

func TestEdgeCommitsResponseBeforeObserver(t *testing.T) {
	admission := &Admission{}
	admission.Open()
	routes, err := NewRoutes(func(error) {}, Route{Method: http.MethodGet,
		Pattern: "/api/items/{id}", Handler: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if err := requestattribution.Bind(r.Context(), [16]byte{1}, [16]byte{2}); err != nil {
					t.Fatal(err)
				}
				w.WriteHeader(http.StatusCreated)
			})})
	if err != nil {
		t.Fatal(err)
	}
	writer := &observationOrderWriter{header: make(http.Header)}
	var captured RequestObservation
	handler, err := NewHandler(admission, routes, func(error) {},
		func(_ context.Context, observation RequestObservation) error {
			if !writer.committed {
				t.Fatal("observer ran before response commit")
			}
			captured = observation
			return errors.New("best-effort observer failed")
		})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/items/private", nil)
	request.Header.Set(requestmeta.CorrelationHeader, "edge-order")
	handler.ServeHTTP(writer, request)
	if writer.status != http.StatusCreated || captured.StatusCode != http.StatusCreated ||
		captured.SourceIdentity == [32]byte{} ||
		captured.AttributionState != requestattribution.StatusVerified ||
		!captured.Attribution.Valid() {
		t.Fatalf("response/observation = %d, %#v", writer.status, captured)
	}
}

func TestEdgeObservationSourceIdentityIsUniquePerRequest(t *testing.T) {
	admission := &Admission{}
	admission.Open()
	routes, err := NewRoutes(func(error) {}, Route{Method: http.MethodGet,
		Pattern: "/api/items", Handler: http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })})
	if err != nil {
		t.Fatal(err)
	}
	identities := make([][32]byte, 0, 2)
	handler, err := NewHandler(admission, routes, func(error) {},
		func(_ context.Context, observation RequestObservation) error {
			identities = append(identities, observation.SourceIdentity)
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	requestEdge(handler, http.MethodGet, "/api/items")
	requestEdge(handler, http.MethodGet, "/api/items")
	if len(identities) != 2 || identities[0] == [32]byte{} || identities[0] == identities[1] {
		t.Fatalf("source identities = %x, %x", identities[0], identities[1])
	}
}
