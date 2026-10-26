package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRequestSearchNormalizesNFC(t *testing.T) {
	search, err := NewRequestSearch("Cafe\u0301")
	if err != nil {
		t.Fatal(err)
	}
	if search.Term() != "Caf\u00e9" {
		t.Fatalf("term = %q", search.Term())
	}
}

func TestRequestSearchEnforcesBounds(t *testing.T) {
	invalid := []string{
		"", " \t\n", "valid\x00term", string([]byte{0xff}),
		strings.Repeat("a", maximumSearchRunes+1),
		strings.Repeat("\u20ac", maximumSearchBytes/utf8.RuneLen('\u20ac')+1),
	}
	for _, term := range invalid {
		if _, err := NewRequestSearch(term); err == nil {
			t.Fatalf("invalid search accepted: bytes=%d runes=%d", len(term), utf8.RuneCountInString(term))
		}
	}
	if _, err := NewRequestSearch(strings.Repeat("a", maximumSearchRunes)); err != nil {
		t.Fatalf("maximum search rejected: %v", err)
	}
}

func TestRequestSearchIsRedactedAndCannotMarshal(t *testing.T) {
	search, _ := NewRequestSearch("secret term")
	for _, value := range []string{fmt.Sprint(search), fmt.Sprintf("%#v", search)} {
		if value != queryValueRedaction {
			t.Fatalf("search formatting leaked: %q", value)
		}
	}
	if _, err := json.Marshal(search); err == nil {
		t.Fatal("opaque search silently marshaled")
	}
}
