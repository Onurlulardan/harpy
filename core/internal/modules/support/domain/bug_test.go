package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func compactJSON(document []byte, maximum int) ([]byte, error) {
	if len(document) == 0 || len(document) > maximum || !json.Valid(document) {
		return nil, fmt.Errorf("invalid JSON")
	}
	var result bytes.Buffer
	if err := json.Compact(&result, document); err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}

func TestNormalizeBugCanonicalizesDiagnosticsAndAppliesDefaults(t *testing.T) {
	input, err := NormalizeBug(BugInput{Title: " Broken page ",
		Description: " The page cannot be opened ", PageURL: "https://example.com/app",
		BrowserInfo:   json.RawMessage(`{"name":"browser"}`),
		ConsoleErrors: json.RawMessage(`[ {"message":"failed"} ]`)}, compactJSON)
	if err != nil {
		t.Fatal(err)
	}
	if input.Title != "Broken page" || input.Category != "bug" ||
		input.Severity != "medium" || string(input.ConsoleErrors) != `[{"message":"failed"}]` {
		t.Fatalf("unexpected normalized bug: %#v", input)
	}
}

func TestNormalizeBugRejectsUnsafeDiagnosticsAndURL(t *testing.T) {
	base := BugInput{Title: "Broken page", Description: "The page cannot be opened"}
	invalidURL := base
	invalidURL.PageURL = "javascript:alert(1)"
	if _, err := NormalizeBug(invalidURL, compactJSON); err == nil {
		t.Fatal("unsafe page URL was accepted")
	}
	tooMany := make([]json.RawMessage, 51)
	for index := range tooMany {
		tooMany[index] = json.RawMessage(`{}`)
	}
	document, err := json.Marshal(tooMany)
	if err != nil {
		t.Fatal(err)
	}
	base.ConsoleErrors = document
	if _, err = NormalizeBug(base, compactJSON); err == nil {
		t.Fatal("over-budget console error list was accepted")
	}
	base.ConsoleErrors = nil
	base.Description = strings.Repeat("x", 8001)
	if _, err = NormalizeBug(base, compactJSON); err == nil {
		t.Fatal("over-budget description was accepted")
	}
}
