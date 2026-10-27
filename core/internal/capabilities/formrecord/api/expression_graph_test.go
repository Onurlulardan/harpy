package recordkernel

import (
	"github.com/google/uuid"
	"testing"
)

func TestDefaultAndCalculationCyclesFailAtCatalogCompile(t *testing.T) {
	leftID, rightID := fixtureV7(401), fixtureV7(402)
	stored := map[uuid.UUID]Field{
		leftID:  {ID: leftID, TypeID: TypeInteger, Role: RoleStored},
		rightID: {ID: rightID, TypeID: TypeInteger, Role: RoleStored},
	}
	leftText, rightText := leftID.String(), rightID.String()
	defaults := []DefaultRule{
		{FieldID: leftText, Expression: typedField("integer", rightText)},
		{FieldID: rightText, Expression: typedField("integer", leftText)},
	}
	if _, _, err := compileDefaults(defaults, stored, expectedHardLimits()); err == nil {
		t.Fatal("default cycle was accepted")
	}
	calculated := map[uuid.UUID]Field{
		leftID: {ID: leftID, TypeID: TypeInteger,
			Role: RoleCalculation},
		rightID: {ID: rightID, TypeID: TypeInteger,
			Role: RoleCalculation},
	}
	calculations := []CalculationDefinition{
		{FieldID: leftText, Expression: Expression{
			Kind: "field", FieldID: &rightText}},
		{FieldID: rightText, Expression: Expression{
			Kind: "field", FieldID: &leftText}},
	}
	if _, _, err := compileCalculations(
		calculations, calculated, expectedHardLimits(),
	); err == nil {
		t.Fatal("calculation cycle was accepted")
	}
}

func TestCrossFieldRulesCompileInStableUUIDOrder(t *testing.T) {
	high, low := fixtureV7(499), fixtureV7(450)
	rules := []CrossFieldValidation{
		{ID: high.String(), Code: "HIGH_RULE",
			Predicate: typedValueLiteral("boolean", `true`)},
		{ID: low.String(), Code: "LOW_RULE",
			Predicate: typedValueLiteral("boolean", `true`)},
	}
	compiled, _, err := compileCrossFieldRules(
		rules, map[uuid.UUID]Field{}, expectedHardLimits(),
	)
	if err != nil || len(compiled) != 2 || compiled[0].id != low ||
		compiled[1].id != high {
		t.Fatalf("stable order = %#v, %v", compiled, err)
	}
}

func TestTypedExpressionDepthLimitFailsClosed(t *testing.T) {
	value := typedValueLiteral("boolean", `true`)
	for index := 0; index < MaximumExpressionDepth; index++ {
		value = typedUnary("not", value)
	}
	if _, _, _, err := compileTypedExpression(value, TypeBoolean,
		map[uuid.UUID]Field{}, 1, expectedHardLimits()); err == nil {
		t.Fatal("over-depth expression was accepted")
	}
}
