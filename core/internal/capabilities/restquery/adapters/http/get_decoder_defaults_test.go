package httpquery

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestGETDecoderCanonicalEmptyDefaults(t *testing.T) {
	request := decodeGETSuccess(t, "")
	if request.Mode() != api.QueryModeRecord || len(request.Select()) != 0 ||
		len(request.Sort()) != 0 || len(request.Includes()) != 0 {
		t.Fatal("empty GET did not produce empty record collections")
	}
	if _, present := request.Where(); present {
		t.Fatal("empty GET gained a filter")
	}
	if _, present := request.Search(); present {
		t.Fatal("empty GET gained search")
	}
	if _, present := request.Aggregate(); present {
		t.Fatal("empty GET gained aggregate state")
	}
	if request.Page().Size() != 50 || request.Count() != api.CountModeNone {
		t.Fatal("empty GET defaults changed")
	}
	if _, present := request.Page().After(); present {
		t.Fatal("empty GET gained a cursor")
	}
	if err := request.Validate(); err != nil {
		t.Fatalf("empty GET request is invalid: %v", err)
	}
}

func TestGETDecoderPageMembersDefaultIndependently(t *testing.T) {
	request := decodeGETSuccess(t, "page[after]=cursor")
	if request.Page().Size() != 50 {
		t.Fatal("missing GET page size did not default")
	}
	if after, present := request.Page().After(); !present || after != "cursor" {
		t.Fatal("GET cursor changed")
	}
	request = decodeGETSuccess(t, "page[size]=25")
	if request.Page().Size() != 25 {
		t.Fatal("GET page size changed")
	}
	if _, present := request.Page().After(); present {
		t.Fatal("missing GET cursor became present")
	}
}
