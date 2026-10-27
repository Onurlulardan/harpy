package main

import (
	"strings"
	"testing"
)

func TestAccessKeysUseUppercaseDigits(t *testing.T) {
	for range 100 {
		key, err := reuseAccessKey("", true)
		if err != nil {
			t.Fatal(err)
		}
		if len(key) != accessKeyLength {
			t.Fatalf("length %d", len(key))
		}
		if strings.EqualFold(key, "rustfsadmin") || validateAccessKey(key) != nil {
			t.Fatal(key)
		}
	}
}

func TestStoredAccessKeyIsKeptUntilRotate(t *testing.T) {
	kept, err := reuseAccessKey("ABC123", false)
	if err != nil || kept != "ABC123" {
		t.Fatalf("kept %q err %v", kept, err)
	}
	if _, err := reuseAccessKey("rustfsadmin", false); err == nil {
		t.Fatal("kept rustfsadmin")
	}
	if _, err := reuseAccessKey("not-upper", false); err == nil {
		t.Fatal("kept invalid access key")
	}
	rotated, err := reuseAccessKey("ABC123", true)
	if err != nil || rotated == "ABC123" {
		t.Fatalf("rotated %q err %v", rotated, err)
	}
}
