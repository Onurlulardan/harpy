package actorbridge

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

func TestForwardInputBindsMethodTargetQueryAndExactReplayBody(t *testing.T) {
	request, _ := http.NewRequest(
		http.MethodPost, "/api/v1/forms/records?sort=id", bytes.NewReader([]byte("first")),
	)
	input, err := NewForwardInput(
		actorassertionv1.DomainDocuments, "documents.items", "create", 7, 19, request,
	)
	if err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*http.Request){
		"method": func(candidate *http.Request) { candidate.Method = http.MethodPut },
		"path":   func(candidate *http.Request) { candidate.URL.Path += "/other" },
		"query":  func(candidate *http.Request) { candidate.URL.RawQuery = "sort=name" },
		"body": func(candidate *http.Request) {
			body := []byte("other")
			candidate.ContentLength = int64(len(body))
			candidate.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(body)), nil
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := request.Clone(request.Context())
			mutate(candidate)
			if _, err := input.validateRequest(candidate); !FaultIs(err, FaultInvalid) {
				t.Fatalf("mutated request error = %v", err)
			}
		})
	}
}

func TestForwardInputRequiresReplayableBoundedBody(t *testing.T) {
	request, _ := http.NewRequest(http.MethodPost, "/records", bytes.NewReader([]byte("body")))
	request.GetBody = nil
	if _, err := NewForwardInput(
		actorassertionv1.DomainDocuments, "documents.items", "create", 1, 1, request,
	); !FaultIs(err, FaultInvalid) {
		t.Fatalf("non-replayable body error = %v", err)
	}
}
