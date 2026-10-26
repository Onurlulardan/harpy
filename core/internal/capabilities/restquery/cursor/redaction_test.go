package cursor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestCursorPublicValuesRedactAllFormattingLoggingAndJSON(t *testing.T) {
	codec, context, token := encodedFixture(t)
	claims, err := codec.Decode(token, context)
	if err != nil {
		t.Fatal(err)
	}
	sorts, _ := context.Sorts()
	last, _ := claims.LastValues()
	values := []any{codec, context, sorts[0], last[0], claims}
	for _, value := range values {
		formatted := fmt.Sprintf("%v|%+v|%#v|%s|%q", value, value, value, value, value)
		for _, secret := range []string{"record-a", testTenantID, "current"} {
			if strings.Contains(formatted, secret) {
				t.Fatalf("format leaked %q from %T", secret, value)
			}
		}
		if !strings.Contains(formatted, redacted) {
			t.Fatalf("format did not redact %T: %s", value, formatted)
		}
		encoded, marshalErr := json.Marshal(value)
		if marshalErr == nil || len(encoded) != 0 {
			t.Fatalf("direct JSON marshal accepted %T: %s", value, encoded)
		}
		var output bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&output, nil))
		logger.Info("cursor", "value", value)
		if strings.Contains(output.String(), "record-a") ||
			strings.Contains(output.String(), testTenantID) ||
			strings.Contains(output.String(), "current") {
			t.Fatalf("structured log leaked %T: %s", value, output.String())
		}
	}
}

func TestClaimsAndContextAccessorsAreMutationIsolated(t *testing.T) {
	codec, context, token := encodedFixture(t)
	claims, err := codec.Decode(token, context)
	if err != nil {
		t.Fatal(err)
	}
	last, err := claims.LastValues()
	if err != nil {
		t.Fatal(err)
	}
	last[0].valid = false
	again, _ := claims.LastValues()
	if again[0].Validate() != nil {
		t.Fatal("claims last-value accessor leaked storage")
	}
	sorts, _ := context.Sorts()
	sorts[0].valid = false
	againSorts, _ := context.Sorts()
	if againSorts[0].Validate() != nil {
		t.Fatal("context sort accessor leaked storage")
	}
}

func TestZeroCursorPublicValuesFailClosed(t *testing.T) {
	if (Codec{}).Validate() == nil || (Context{}).Validate() == nil ||
		(SortDefinition{}).Validate() == nil || (LastValue{}).Validate() == nil ||
		(Claims{}).Validate() == nil {
		t.Fatal("zero cursor value validated")
	}
}
