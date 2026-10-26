package postgres

import (
	"slices"
	"testing"
)

func TestV1RegistryContractKeepsExactCompatibilitySurface(t *testing.T) {
	wantColumns := []string{
		"id", "name", "slug", "status", "settings", "owner_id",
		"created_at", "updated_at", "deleted_at",
	}
	if !slices.Equal(V1RegistryColumns[:], wantColumns) {
		t.Fatalf("V1 columns = %v, want %v", V1RegistryColumns, wantColumns)
	}
	wantWritable := []string{
		"name", "status", "settings", "owner_id", "updated_at", "deleted_at",
	}
	if !slices.Equal(V1RegistryWritableColumns[:], wantWritable) {
		t.Fatalf("V1 writable columns = %v, want %v", V1RegistryWritableColumns, wantWritable)
	}
}
