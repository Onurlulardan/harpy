package schemaexecutor

import (
	"strings"
	"testing"
)

func TestDeriveObjectIdentifierUsesLowercaseRFC4648(t *testing.T) {
	actual, err := deriveObjectIdentifier(
		"frm_", "00112233-4455-6677-8899-aabbccddeeff",
	)
	if err != nil {
		t.Fatal(err)
	}
	const expected = "frm_aaisem2ekvthpcezvk54zxpo74"
	if actual != expected {
		t.Fatalf("identifier = %q, want %q", actual, expected)
	}
}

func TestIdentifierMappingRejectsUppercasePhysicalName(t *testing.T) {
	plan := testCreateFormPlan(t)
	plan.Steps[0].IdentifierMapping.Form.PhysicalName = strings.ToUpper(
		plan.Steps[0].IdentifierMapping.Form.PhysicalName,
	)
	if _, err := BuildPlan(plan); err == nil {
		t.Fatal("uppercase physical identifier was accepted")
	}
}

func FuzzDeriveObjectIdentifierNeverReturnsUppercase(f *testing.F) {
	f.Add("00112233-4455-6677-8899-aabbccddeeff")
	f.Fuzz(func(t *testing.T, value string) {
		identifier, err := deriveObjectIdentifier("idx_", value)
		if err == nil && identifier != strings.ToLower(identifier) {
			t.Fatalf("identifier contains uppercase: %q", identifier)
		}
	})
}
