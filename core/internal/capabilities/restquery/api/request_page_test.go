package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestRequestPageValidatesSizeAndOptionalCursor(t *testing.T) {
	page, err := NewRequestPage(50, nil)
	if err != nil {
		t.Fatal(err)
	}
	if page.Size() != 50 {
		t.Fatalf("size = %d", page.Size())
	}
	if _, found := page.After(); found {
		t.Fatal("absent cursor became present")
	}
	cursor := "opaque-cursor"
	page, err = NewRequestPage(500, &cursor)
	if err != nil {
		t.Fatal(err)
	}
	cursor = "changed"
	if after, found := page.After(); !found || after != "opaque-cursor" {
		t.Fatalf("cursor = %q, found = %v", after, found)
	}
}

func TestRequestPageRejectsInvalidBounds(t *testing.T) {
	for _, size := range []int{0, 501} {
		if _, err := NewRequestPage(size, nil); err == nil {
			t.Fatalf("size %d accepted", size)
		}
	}
	invalidUTF8 := string([]byte{0xff})
	if _, err := NewRequestPage(1, &invalidUTF8); err == nil {
		t.Fatal("invalid UTF-8 cursor accepted")
	}
	empty := ""
	if _, err := NewRequestPage(1, &empty); err == nil {
		t.Fatal("present empty cursor accepted")
	}
	tooLong := strings.Repeat("x", maximumCursorBytes+1)
	if _, err := NewRequestPage(1, &tooLong); err == nil {
		t.Fatal("oversized cursor accepted")
	}
}

func TestRequestPageIsRedactedAndCannotMarshal(t *testing.T) {
	cursor := "secret-cursor"
	page, _ := NewRequestPage(50, &cursor)
	for _, value := range []string{fmt.Sprint(page), fmt.Sprintf("%#v", page)} {
		if value != queryValueRedaction {
			t.Fatalf("page formatting leaked: %q", value)
		}
	}
	if _, err := json.Marshal(page); err == nil {
		t.Fatal("opaque page silently marshaled")
	}
}
