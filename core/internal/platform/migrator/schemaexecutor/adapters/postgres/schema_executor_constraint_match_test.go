package postgres

import "testing"

func TestNormalizeConstraintDefinitionSeparatesValidationState(t *testing.T) {
	actual := schemaExecutorNormalizeDefinition(`CHECK (("field_g1" IS NOT NULL)) NOT VALID`)
	if actual != "checkfield_g1isnotnull" {
		t.Fatalf("normalized constraint definition = %s", actual)
	}
}
