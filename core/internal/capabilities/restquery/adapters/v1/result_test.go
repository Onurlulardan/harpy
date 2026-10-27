package v1query

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestV1ResultIsCanonicalRecordQuery(t *testing.T) {
	result, err := mustDecoder(t).Decode("")
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	request, err := result.Request()
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}
	if request.Mode() != api.QueryModeRecord || request.Count() != api.CountModeExact {
		t.Fatalf("mode/count = %v/%v", request.Mode(), request.Count())
	}
	if len(request.Select()) != 0 || len(request.Includes()) != 0 {
		t.Fatalf("select/includes = %#v/%#v", request.Select(), request.Includes())
	}
	if len(request.Sort()) != 1 || request.Sort()[0].Field() != "created_at" ||
		request.Sort()[0].Direction() != api.DirectionDescending {
		t.Fatalf("default sort = %#v", request.Sort())
	}
	if _, found := request.Where(); found {
		t.Fatal("default request unexpectedly contains a filter")
	}
	if _, found := request.Search(); found {
		t.Fatal("default request unexpectedly contains search")
	}
}

func TestV1AdapterValuesAreRedactedAndRejectDirectJSON(t *testing.T) {
	decoder := mustDecoder(t)
	result, err := decoder.Decode("search=secret")
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	plan, _ := result.PagePlan()
	values := []any{decoder, testManifest(), result, plan}
	for _, value := range values {
		if actual := fmt.Sprintf("%v", value); actual != v1Redaction {
			t.Fatalf("formatted value = %q, want redaction", actual)
		}
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T unexpectedly marshaled to JSON", value)
		}
	}
}

func TestZeroValueAdapterResultsFailValidation(t *testing.T) {
	if err := (decoder{}).Validate(); err == nil {
		t.Fatal("zero Decoder unexpectedly validated")
	}
	if err := (pendingResult{}).Validate(); err == nil {
		t.Fatal("zero Result unexpectedly validated")
	}
	if err := (PagePlan{}).Validate(); err == nil {
		t.Fatal("zero PagePlan unexpectedly validated")
	}
}
