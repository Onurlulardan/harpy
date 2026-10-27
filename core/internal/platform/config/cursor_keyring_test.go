package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestCursorKeyringFormattingAndJSONAlwaysRedact(t *testing.T) {
	keyring, err := parseCursorKeyring(Secret{value: validCursorKeyringJSON()})
	if err != nil {
		t.Fatal(err)
	}
	current, err := keyring.Current()
	if err != nil {
		t.Fatal(err)
	}
	values := []any{keyring, current}
	for _, value := range values {
		for _, rendered := range []string{
			fmt.Sprint(value), fmt.Sprintf("%s", value), fmt.Sprintf("%v", value),
			fmt.Sprintf("%+v", value), fmt.Sprintf("%#v", value), fmt.Sprintf("%q", value),
		} {
			assertCursorRedacted(t, rendered)
		}
		document, marshalErr := json.Marshal(value)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		assertCursorRedacted(t, string(document))
	}
}

func TestCursorKeyringLogValueAlwaysRedacts(t *testing.T) {
	keyring, err := parseCursorKeyring(Secret{value: validCursorKeyringJSON()})
	if err != nil {
		t.Fatal(err)
	}
	current, err := keyring.Current()
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	logger.Info("config", "keyring", keyring, "key", current)
	assertCursorRedacted(t, output.String())
	if strings.Count(output.String(), cursorRedacted) != 2 {
		t.Fatalf("cursor values were not both redacted: %s", output.String())
	}
}

func assertCursorRedacted(t *testing.T, rendered string) {
	t.Helper()
	if strings.Contains(rendered, cursorTestKey(1)) ||
		strings.Contains(rendered, cursorTestKey(2)) ||
		strings.Contains(rendered, "current-1") ||
		!strings.Contains(rendered, cursorRedacted) {
		t.Fatalf("cursor secret leaked or marker missing: %q", rendered)
	}
}
