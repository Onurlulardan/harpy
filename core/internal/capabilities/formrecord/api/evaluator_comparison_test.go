package recordkernel

import (
	"testing"
)

func TestComparisonOperatorMatrix(t *testing.T) {
	dateEarly, _ := decodeTemporal(TypeDate, []byte(`"2026-01-01"`))
	dateLate, _ := decodeTemporal(TypeDate, []byte(`"2026-01-02"`))
	timeEarly, _ := decodeTemporal(TypeTime, []byte(`"10:00:00.000000"`))
	timeLate, _ := decodeTemporal(TypeTime, []byte(`"10:00:00.000001"`))
	timestampEarly, _ := decodeTemporal(TypeTimestamp,
		[]byte(`"2026-01-01T00:00:00.000000Z"`))
	timestampLate, _ := decodeTemporal(TypeTimestamp,
		[]byte(`"2026-01-02T00:00:00.000000Z"`))
	decimalOne, _ := decodeGenericDecimal([]byte(`"1.0"`))
	decimalTwo, _ := decodeGenericDecimal([]byte(`"2.00"`))
	textA, _ := buildText("a\nb")
	textB, _ := buildText("a\"b")
	tests := []struct {
		name     string
		operator string
		left     Value
		right    Value
	}{
		{"integer lt", "lt", buildInteger(1), buildInteger(2)},
		{"decimal lt", "lt", decimalOne, decimalTwo},
		{"boolean eq", "eq", buildBoolean(true), buildBoolean(true)},
		{"text decoded lt", "lt", textA, textB},
		{"date lt", "lt", dateEarly, dateLate},
		{"time lt", "lt", timeEarly, timeLate},
		{"timestamp lt", "lt", timestampEarly, timestampLate},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			typeID := test.left.TypeID()
			expression := expression{kind: "binary", typeID: TypeBoolean,
				operator: test.operator, arguments: []expression{
					{kind: "literal", typeID: typeID, literal: test.left},
					{kind: "literal", typeID: typeID, literal: test.right},
				}}
			value, err := evaluateTestExpression(expression, nil)
			if err != nil || string(value.CanonicalJSON()) != "true" {
				t.Fatalf("comparison = %s, %v", value.CanonicalJSON(), err)
			}
		})
	}
}

func TestComparisonEqualityAndNullSemantics(t *testing.T) {
	decomposed, _ := buildText("e\u0301")
	composed, _ := buildText("é")
	for _, operator := range []string{"eq", "gte", "lte"} {
		expression := expression{kind: "binary", typeID: TypeBoolean,
			operator: operator, arguments: []expression{
				{kind: "literal", typeID: TypeText, literal: decomposed},
				{kind: "literal", typeID: TypeText, literal: composed},
			}}
		value, err := evaluateTestExpression(expression, nil)
		if err != nil || string(value.CanonicalJSON()) != "true" {
			t.Fatalf("NFC %s = %s, %v", operator, value.CanonicalJSON(), err)
		}
	}
	nullEquality := expression{kind: "binary", typeID: TypeBoolean,
		operator: "eq", arguments: []expression{
			{kind: "literal", typeID: TypeText,
				literal: nullValue(TypeText)},
			textLiteral(t, "value"),
		}}
	value, err := evaluateTestExpression(nullEquality, nil)
	if err != nil || !value.IsNull() {
		t.Fatalf("null equality = %+v, %v", value, err)
	}
}
