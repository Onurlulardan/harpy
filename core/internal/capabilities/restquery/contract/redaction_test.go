package contract

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestRawMemberAndQueryDocumentAreRedacted(t *testing.T) {
	t.Parallel()
	member := mustRawMember(t, `{"password":"super-secret"}`)
	document, err := NewQueryDocument(QueryDocumentInput{Where: member})
	if err != nil {
		t.Fatal(err)
	}
	values := []any{member, document}
	for _, value := range values {
		outputs := []string{
			fmt.Sprintf("%s", value), fmt.Sprintf("%q", value),
			fmt.Sprintf("%v", value), fmt.Sprintf("%+v", value),
			fmt.Sprintf("%#v", value), fmt.Sprintf("%x", value),
		}
		for _, output := range outputs {
			if output == "" || strings.Contains(output, "super-secret") {
				t.Fatalf("format exposed raw JSON: %q", output)
			}
		}
		encoded, marshalErr := json.Marshal(value)
		if marshalErr == nil {
			t.Fatalf("value crossed JSON boundary: %s", encoded)
		}
		if strings.Contains(marshalErr.Error(), "super-secret") {
			t.Fatalf("marshal error exposed raw JSON: %v", marshalErr)
		}
	}
}
