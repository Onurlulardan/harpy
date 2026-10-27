package recordkernel

import (
	"encoding/json"
	"github.com/google/uuid"
	"testing"
)

func TestTypedExpressionClosedSignatureMatrix(t *testing.T) {
	fields := expressionFixtureFields()
	textID := fixtureV7(301).String()
	valid := []struct {
		expected TypeID
		value    TypedExpression
	}{
		{TypeInteger, typedUnary("negate", typedValueLiteral("integer", `"1"`))},
		{TypeBoolean, typedUnary("not", typedValueLiteral("boolean", `true`))},
		{TypeInteger, typedBinary("integer", "add",
			typedValueLiteral("integer", `"1"`), typedValueLiteral("integer", `"2"`))},
		{TypeBoolean, typedBinary("boolean", "eq",
			typedField("text", textID), typedValueLiteral("text", `"a"`))},
		{TypeText, typedCall("text", "coalesce",
			typedField("text", textID), typedValueLiteral("text", `"x"`))},
		{TypeText, typedCall("text", "concat",
			typedValueLiteral("text", `"a"`), typedValueLiteral("text", `"b"`))},
		{TypeText, typedCall("text", "lower", typedValueLiteral("text", `"A"`))},
		{TypeText, typedCall("text", "upper", typedValueLiteral("text", `"a"`))},
		{TypeDecimal, typedCall("decimal", "round",
			typedValueLiteral("decimal", `"1.25"`), typedValueLiteral("integer", `"1"`))},
		{TypeDate, typedCall("date", "date_add",
			typedValueLiteral("date", `"2026-01-01"`),
			typedValueLiteral("integer", `"1"`))},
	}
	for index, test := range valid {
		if _, _, _, err := compileTypedExpression(
			test.value, test.expected, fields, 1, expectedHardLimits(),
		); err != nil {
			t.Fatalf("valid signature %d: %v", index, err)
		}
	}
	invalid := []TypedExpression{
		typedBinary("text", "add", typedValueLiteral("text", `"a"`),
			typedValueLiteral("text", `"b"`)),
		typedBinary("boolean", "eq", typedValueLiteral("integer", `"1"`),
			typedValueLiteral("text", `"1"`)),
		typedCall("decimal", "round", typedValueLiteral("decimal", `"1"`)),
		typedCall("time", "date_add", typedValueLiteral("time", `"10:00:00.000000"`),
			typedValueLiteral("integer", `"1"`)),
		typedCall("text", "unknown", typedValueLiteral("text", `"x"`)),
	}
	for index, value := range invalid {
		if _, _, _, err := compileTypedExpression(
			value, TypeID(value.Type), fields, 1, expectedHardLimits(),
		); err == nil {
			t.Fatalf("invalid signature %d accepted", index)
		}
	}
}

func TestLegacyExpressionInferenceFailsClosedWithoutFieldHint(t *testing.T) {
	fields := expressionFixtureFields()
	add := "add"
	valid := Expression{Kind: "binary", Operator: &add,
		Arguments: []Expression{{Kind: "literal", Literal: json.RawMessage(`1`)},
			{Kind: "literal", Literal: json.RawMessage(`2`)}}}
	if _, _, _, err := compileLegacyExpression(
		valid, TypeInteger, fields, 1, expectedHardLimits(),
	); err != nil {
		t.Fatal(err)
	}
	eq := "eq"
	ambiguous := Expression{Kind: "binary", Operator: &eq,
		Arguments: []Expression{{Kind: "literal", Literal: json.RawMessage(`1`)},
			{Kind: "literal", Literal: json.RawMessage(`1`)}}}
	if _, _, _, err := compileLegacyExpression(
		ambiguous, TypeBoolean, fields, 1, expectedHardLimits(),
	); err == nil {
		t.Fatal("ambiguous legacy comparison was accepted")
	}
	fieldID := fixtureV7(302).String()
	hinted := ambiguous
	hinted.Arguments[0] = Expression{Kind: "field", FieldID: &fieldID}
	if _, _, _, err := compileLegacyExpression(
		hinted, TypeBoolean, fields, 1, expectedHardLimits(),
	); err != nil {
		t.Fatalf("field-hinted comparison: %v", err)
	}
}

func expressionFixtureFields() map[uuid.UUID]Field {
	return map[uuid.UUID]Field{
		fixtureV7(301): {ID: fixtureV7(301), TypeID: TypeText,
			Role: RoleStored},
		fixtureV7(302): {ID: fixtureV7(302), TypeID: TypeInteger,
			Role: RoleStored},
	}
}

func typedValueLiteral(typeID string, raw string) TypedExpression {
	return TypedExpression{Kind: "literal", Type: typeID, Value: json.RawMessage(raw)}
}

func typedField(typeID, fieldID string) TypedExpression {
	return TypedExpression{Kind: "field", Type: typeID, FieldID: &fieldID}
}

func typedUnary(operator string, argument TypedExpression) TypedExpression {
	return TypedExpression{Kind: "unary", Type: argument.Type,
		Operator: &operator, Arguments: []TypedExpression{argument}}
}

func typedBinary(
	typeID, operator string, left, right TypedExpression,
) TypedExpression {
	return TypedExpression{Kind: "binary", Type: typeID, Operator: &operator,
		Arguments: []TypedExpression{left, right}}
}

func typedCall(typeID, function string, arguments ...TypedExpression) TypedExpression {
	return TypedExpression{Kind: "call", Type: typeID, Function: &function,
		Arguments: arguments}
}
