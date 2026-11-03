package domain

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
)

func TestSchemaNameMatchesFoundationTenantNamespace(t *testing.T) {
	id := TenantID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	got, err := SchemaNameForTenant(id)
	if err != nil {
		t.Fatal(err)
	}
	want, err := foundation.TenantSchemaName(id)
	if err != nil || got.String() != want {
		t.Fatalf("schema = %s want %s err %v", got.String(), want, err)
	}
	if _, err := SchemaNameForTenant(TenantID{}); err == nil {
		t.Fatal("accepted the zero tenant ID")
	}
}
