package work

import (
	"testing"

	workhttp "github.com/Onurlulardan/harpy/core/internal/modules/work/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestLegacyTaskPriorityIsValidatedAndForwarded(t *testing.T) {
	value, err := legacyReadOptions(workhttp.ListInput{Compatibility: true,
		Resource: "inbox", LegacyQuery: map[string]string{"priority": "URGENT"}})
	if err != nil || value.Priority != ports.PriorityUrgent {
		t.Fatalf("priority filter lost: %#v %v", value, err)
	}
	_, err = legacyReadOptions(workhttp.ListInput{Compatibility: true,
		Resource: "inbox", LegacyQuery: map[string]string{"priority": "UNKNOWN"}})
	if err == nil {
		t.Fatal("unknown priority accepted")
	}
}
