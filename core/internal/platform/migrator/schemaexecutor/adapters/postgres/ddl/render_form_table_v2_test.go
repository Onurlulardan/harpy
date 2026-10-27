package ddl

import (
	"strings"
	"testing"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func TestV2EmptyArgumentsUseDecodedValueShape(t *testing.T) {
	if err := validateEmptyArgumentsV2(schemav2.EmptyArguments{}); err != nil {
		t.Fatalf("decoded V2 empty arguments rejected: %v", err)
	}
	if err := validateEmptyArgumentsV2(&schemav2.EmptyArguments{}); err == nil {
		t.Fatal("pointer arguments accepted as exact V2 value shape")
	}
}

func TestV2FormTablePinsIdentityAuditAndTimeGuards(t *testing.T) {
	checks := formTableChecksV2("frm_contract")
	for _, contract := range []string{
		"uuid_extract_version(id) IS NOT DISTINCT FROM 7",
		"created_by <> '00000000-0000-0000-0000-000000000000'::uuid",
		"updated_by <> '00000000-0000-0000-0000-000000000000'::uuid",
		"deleted_by IS NULL OR deleted_by <> '00000000-0000-0000-0000-000000000000'::uuid",
		"pg_catalog.isfinite(created_at)", "pg_catalog.isfinite(updated_at)",
		"created_at <= updated_at", "pg_catalog.isfinite(deleted_at)",
		"deleted_at >= created_at",
		"deleted_at <= updated_at",
		"(deleted_at IS NULL) = (deleted_by IS NULL)",
	} {
		if !strings.Contains(checks, contract) {
			t.Fatalf("V2 form table is missing hard guard %q", contract)
		}
	}
}
