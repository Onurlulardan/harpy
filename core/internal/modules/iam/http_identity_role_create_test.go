package iam

import (
	"testing"

	"github.com/google/uuid"
)

func TestIdentityRoleScopeInputAcceptsV1InferenceAndExplicitV2Scope(t *testing.T) {
	tenantID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	organizationID := uuid.MustParse("20000000-0000-4000-8000-000000000001")
	tests := []struct {
		organizationID, wantID uuid.UUID
		kind, raw, wantKind    string
	}{
		{uuid.Nil, tenantID, "", "", "TENANT"},
		{organizationID, organizationID, "", "", "ORGANIZATION"},
		{organizationID, organizationID, "ORGANIZATION", organizationID.String(), "ORGANIZATION"},
	}
	for _, test := range tests {
		kind, id, valid := identityRoleScopeInput(
			tenantID, test.organizationID, test.kind, test.raw,
		)
		if !valid || kind != test.wantKind || id != test.wantID {
			t.Fatalf("scope=(%q,%s,%t), want=(%q,%s,true)",
				kind, id, valid, test.wantKind, test.wantID)
		}
	}
	if _, _, valid := identityRoleScopeInput(
		tenantID, uuid.Nil, "TENANT", "",
	); valid {
		t.Fatal("partial explicit scope was accepted")
	}
}
