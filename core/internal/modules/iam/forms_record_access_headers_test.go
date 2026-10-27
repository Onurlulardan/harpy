package iam

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestFormsRecordAccessAuthenticationHeadersFailClosed(t *testing.T) {
	for _, values := range [][]string{
		nil,
		{""},
		{"bearer token"},
		{"Bearer "},
		{"Bearer token with-space"},
		{"Bearer first", "Bearer second"},
	} {
		if _, valid := formsRecordBearer(values); valid {
			t.Fatalf("invalid authorization header accepted: %q", values)
		}
	}
	if token, valid := formsRecordBearer([]string{"Bearer exact-token"}); !valid || token != "exact-token" {
		t.Fatalf("valid bearer rejected: %q %t", token, valid)
	}
	if _, err := (*FormsRecordAccessRuntime)(nil).Verify(
		context.Background(), FormsRecordAccessRequest{},
	); err != ErrFormsRecordAccessUnavailable {
		t.Fatalf("nil runtime error = %v", err)
	}
}

func TestFormsRecordAccessRejectsSecondTenantHeader(t *testing.T) {
	tenantA := uuid.MustParse("018f47f4-7b1b-7000-8000-000000000001")
	tenantB := uuid.MustParse("018f47f4-7b1b-7000-8000-000000000002")
	runtime := &FormsRecordAccessRuntime{}
	if runtime.tenantHeadersMatch(context.Background(), FormsRecordAccessRequest{
		TenantIDValues: []string{tenantB.String()},
	}, tenantA) {
		t.Fatal("tenant B header selected tenant A token authority")
	}
	if !runtime.tenantHeadersMatch(context.Background(), FormsRecordAccessRequest{
		TenantIDValues: []string{tenantA.String()},
	}, tenantA) {
		t.Fatal("matching tenant evidence was rejected")
	}
	if runtime.tenantHeadersMatch(context.Background(), FormsRecordAccessRequest{
		TenantIDValues: []string{tenantA.String(), tenantB.String()},
	}, tenantA) {
		t.Fatal("duplicate tenant evidence was accepted")
	}
}
