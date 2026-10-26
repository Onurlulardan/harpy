package api

import (
	"io"
	"strings"
	"testing"
)

func TestCSVReaderIsBoundedAndPreservesColumns(t *testing.T) {
	reader, err := NewReader(FormatCSV, strings.NewReader("email,role_id\na@b.test,r1\n"),
		DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	row, err := reader.Next()
	if err != nil || row.Index != 1 || row.Values["email"] != "a@b.test" {
		t.Fatalf("row=%+v err=%v", row, err)
	}
	if _, err = reader.Next(); err != io.EOF {
		t.Fatalf("terminal error=%v", err)
	}
}

func TestJSONLReaderRejectsShapeDrift(t *testing.T) {
	reader, err := NewReader(FormatJSONL,
		strings.NewReader("{\"email\":\"a@b.test\",\"role_id\":\"r1\"}\n{\"email\":\"b@b.test\"}\n"),
		DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = reader.Next(); err != nil {
		t.Fatal(err)
	}
	if _, err = reader.Next(); err != ErrFormat {
		t.Fatalf("shape drift error=%v", err)
	}
}

func TestTenantPolicyCanOnlyLowerLimits(t *testing.T) {
	limits := DefaultLimits()
	limits.Rows++
	if err := limits.Validate(); err == nil {
		t.Fatal("raised tenant limit was accepted")
	}
}
