package httpquery

import (
	"strings"
	"testing"
)

func TestScanJSONMembersRejectsInvalidDocuments(t *testing.T) {
	tests := []struct {
		name     string
		document []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
		{"whitespace", []byte(" \r\n\t")},
		{"malformed object", []byte(`{"name":}`)},
		{"unfinished object", []byte(`{"name":"value"`)},
		{"invalid utf8", []byte{'"', 0xff, '"'}},
		{"trailing object", []byte(`{} {}`)},
		{"trailing scalar", []byte(`null true`)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := scanJSONMembers(test.document); err == nil {
				t.Fatal("invalid JSON document accepted")
			}
		})
	}
}

func TestScanJSONMembersEnforcesNestingLimit(t *testing.T) {
	if _, err := scanJSONMembers(nestedJSONArray(maximumJSONMemberNesting)); err != nil {
		t.Fatalf("maximum nesting rejected: %v", err)
	}
	if _, err := scanJSONMembers(nestedJSONArray(maximumJSONMemberNesting + 1)); err == nil {
		t.Fatal("over-limit nesting accepted")
	}
}

func TestScanJSONMembersHandles64KiBDocument(t *testing.T) {
	const maximumDocumentBytes = 64 * 1024
	wrapperBytes := len(`{"value":""}`)
	document := []byte(`{"value":"` +
		strings.Repeat("x", maximumDocumentBytes-wrapperBytes) + `"}`)
	if len(document) != maximumDocumentBytes {
		t.Fatalf("fixture size = %d", len(document))
	}
	if _, err := scanJSONMembers(document); err != nil {
		t.Fatalf("64 KiB document rejected: %v", err)
	}
}

func TestScanJSONMembersErrorsDoNotExposeRawMembers(t *testing.T) {
	const secretKey = "super-secret/key"
	const secretValue = "super-secret-value"
	document := []byte(`{"super-secret/key":"super-secret-value",` +
		`"super-secret/key":"other"} {}`)
	_, err := scanJSONMembers(document)
	if err == nil {
		t.Fatal("invalid trailing document accepted")
	}
	if strings.Contains(err.Error(), secretKey) || strings.Contains(err.Error(), secretValue) {
		t.Fatalf("scanner error exposed raw member data: %v", err)
	}
}

func nestedJSONArray(depth int) []byte {
	return []byte(strings.Repeat("[", depth) + "0" + strings.Repeat("]", depth))
}
