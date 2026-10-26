package api

import "testing"

func TestOnboardingSchemaNamePublishesValidatedPlanName(t *testing.T) {
	value := UUID{15: 1}
	name, err := OnboardingSchemaName(value)
	if err != nil || name != onboardingSchemaName(value) {
		t.Fatalf("schema name = %q, %v", name, err)
	}
	if _, err := OnboardingSchemaName(NilUUID); err == nil {
		t.Fatal("nil tenant identity produced a schema name")
	}
}
