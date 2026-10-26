package iam

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestIdentityIdempotencyKeyIsExactAndUnambiguous(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		valid  bool
	}{
		{"minimum", []string{"a234567890123456"}, true},
		{"full alphabet", []string{"A2345678901234._:-"}, true},
		{"too short", []string{"a23456789012345"}, false},
		{"too long", []string{"a" + strings.Repeat("b", 128)}, false},
		{"space", []string{"a23456789012345 "}, false},
		{"duplicate", []string{"a234567890123456", "a234567890123456"}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/api/users", nil)
			for _, value := range test.values {
				request.Header.Add("Idempotency-Key", value)
			}
			_, valid := identityIdempotencyKey(request)
			if valid != test.valid {
				t.Fatalf("valid=%t want=%t", valid, test.valid)
			}
		})
	}
}

func TestIdentityRevisionAndPathHeadersAreCanonical(t *testing.T) {
	id := uuid.MustParse("66abcdef-abcd-4abc-8abc-abcdefabcdef")
	tests := []struct {
		etag  string
		valid bool
	}{
		{`"revision-1"`, true}, {`"revision-9223372036854775807"`, true},
		{`revision-1`, false}, {`"revision-0"`, false}, {`"revision-01"`, false},
		{`W/"revision-1"`, false}, {`"revision-9223372036854775808"`, false},
	}
	for _, test := range tests {
		request := httptest.NewRequest("PUT", "/api/users/"+id.String(), nil)
		request.SetPathValue("id", id.String())
		request.Header.Set("If-Match", test.etag)
		_, valid := identityExpectedRevision(request)
		if valid != test.valid {
			t.Fatalf("If-Match %q valid=%t want=%t", test.etag, valid, test.valid)
		}
	}
	request := httptest.NewRequest("GET", "/api/users/ignored", nil)
	request.SetPathValue("id", strings.ToUpper(id.String()))
	if _, valid := identityPathID(request); valid {
		t.Fatal("non-canonical uppercase UUID was accepted")
	}
}
