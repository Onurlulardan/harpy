package systemresource

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestSensitivePublicValuesAreRedacted(t *testing.T) {
	evidence, _ := NewActorEvidence(
		mustSystemResourceID(t, "018f0f4e-7b61-7c25-8000-000000000061"),
		mustSystemResourceID(t, "018f0f4e-7b61-7c25-8000-000000000062"), 1, 1,
	)
	fieldID := mustSystemResourceID(t, "a1000000-0000-4000-8000-000000000002")
	values, _ := NewFieldValues(map[ID]json.RawMessage{
		fieldID: []byte(`"hidden"`),
	})
	descriptor, _ := NewFieldDescriptorWithCapabilities(
		fieldID, "email", true, true,
	)
	created, _ := NewCreateInput(Users, values, []byte("secret"), "redact-1")
	defer created.Destroy()
	for _, value := range []any{evidence, values, descriptor, created} {
		if rendered := fmt.Sprintf("%+v", value); rendered != redactedValue {
			t.Fatalf("value rendered as %q", rendered)
		}
		if document, err := json.Marshal(value); err == nil ||
			(strings.Contains(string(document), "tenant") ||
				strings.Contains(string(document), "secret")) {
			t.Fatalf("direct JSON=%s error=%v", document, err)
		}
	}
}
