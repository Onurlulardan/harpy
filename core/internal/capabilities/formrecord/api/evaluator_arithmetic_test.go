package recordkernel

import (
	"errors"
	"math"
	"testing"
)

func TestIntegerArithmeticExactMatrix(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		left     int64
		right    int64
		expected string
		fails    bool
	}{
		{"add", "add", 4, 3, `"7"`, false},
		{"subtract", "subtract", 4, 3, `"1"`, false},
		{"multiply", "multiply", -4, 3, `"-12"`, false},
		{"divide", "divide", 12, 3, `"4"`, false},
		{"add overflow", "add", math.MaxInt64, 1, "", true},
		{"subtract overflow", "subtract", math.MinInt64, 1, "", true},
		{"multiply overflow", "multiply", math.MaxInt64, 2, "", true},
		{"non integral divide", "divide", 5, 2, "", true},
		{"divide zero", "divide", 5, 0, "", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expression := binaryInteger(test.operator, test.left, test.right)
			value, err := evaluateTestExpression(expression, nil)
			if test.fails {
				if !errors.Is(err, ErrCalculation) {
					t.Fatalf("err = %v", err)
				}
				return
			}
			if err != nil || string(value.CanonicalJSON()) != test.expected {
				t.Fatalf("value = %s, err = %v", value.CanonicalJSON(), err)
			}
		})
	}
}

func TestNegateBooleanAndStrictNullMatrix(t *testing.T) {
	negate := expression{kind: "unary", typeID: TypeInteger,
		operator: "negate", arguments: []expression{integerLiteral(math.MinInt64)}}
	if _, err := evaluateTestExpression(negate, nil); !errors.Is(err, ErrCalculation) {
		t.Fatalf("min negate err = %v", err)
	}
	negate.arguments[0] = integerLiteral(3)
	value, err := evaluateTestExpression(negate, nil)
	if err != nil || string(value.CanonicalJSON()) != `"-3"` {
		t.Fatalf("negate = %s, %v", value.CanonicalJSON(), err)
	}
	not := expression{kind: "unary", typeID: TypeBoolean,
		operator: "not", arguments: []expression{booleanLiteral(true)}}
	value, err = evaluateTestExpression(not, nil)
	if err != nil || string(value.CanonicalJSON()) != "false" {
		t.Fatalf("not = %s, %v", value.CanonicalJSON(), err)
	}
	for _, test := range []struct {
		operator string
		left     bool
		right    bool
		expected string
	}{{"and", true, false, "false"}, {"or", true, false, "true"}} {
		value, err := evaluateTestExpression(expression{kind: "binary",
			typeID: TypeBoolean, operator: test.operator,
			arguments: []expression{booleanLiteral(test.left), booleanLiteral(test.right)}}, nil)
		if err != nil || string(value.CanonicalJSON()) != test.expected {
			t.Fatalf("%s = %s, %v", test.operator, value.CanonicalJSON(), err)
		}
	}
	nullAnd := expression{kind: "binary", typeID: TypeBoolean,
		operator: "and", arguments: []expression{
			{kind: "literal", typeID: TypeBoolean,
				literal: nullValue(TypeBoolean)},
			booleanLiteral(false),
		}}
	value, err = evaluateTestExpression(nullAnd, nil)
	if err != nil || !value.IsNull() {
		t.Fatalf("strict boolean null = %+v, %v", value, err)
	}
}

func TestDecimalArithmeticRoundsEveryNodeToTargetProfile(t *testing.T) {
	profile := &DecimalProfile{Precision: 6, Scale: 2}
	tests := []struct {
		operator string
		left     string
		right    string
		expected string
		fails    bool
	}{
		{"add", "1.005", "0", `"1.00"`, false},
		{"subtract", "5.555", "1.111", `"4.44"`, false},
		{"multiply", "2.555", "2", `"5.11"`, false},
		{"divide", "1", "8", `"0.12"`, false},
		{"divide", "1", "0", "", true},
		{"multiply", "9999.99", "10", "", true},
	}
	for _, test := range tests {
		left := decimalLiteral(t, test.left)
		right := decimalLiteral(t, test.right)
		value, err := evaluateTestExpression(expression{kind: "binary",
			typeID: TypeDecimal, operator: test.operator,
			arguments: []expression{left, right}}, profile)
		if test.fails {
			if err == nil {
				t.Fatalf("%s %s/%s accepted", test.operator, test.left, test.right)
			}
			continue
		}
		if err != nil || string(value.CanonicalJSON()) != test.expected {
			t.Fatalf("%s = %s, %v", test.operator, value.CanonicalJSON(), err)
		}
	}
	negate := expression{kind: "unary", typeID: TypeDecimal,
		operator: "negate", arguments: []expression{decimalLiteral(t, "1.235")}}
	value, err := evaluateTestExpression(negate, profile)
	if err != nil || string(value.CanonicalJSON()) != `"-1.24"` {
		t.Fatalf("decimal negate = %s, %v", value.CanonicalJSON(), err)
	}
}

func evaluateTestExpression(
	value expression,
	profile *DecimalProfile,
) (Value, error) {
	catalog := fixtureCatalog(nil)
	budget := newInstructionBudget(catalog)
	return evaluateExpression(value, Record{}, profile, &budget)
}

func binaryInteger(operator string, left, right int64) expression {
	return expression{kind: "binary", typeID: TypeInteger,
		operator: operator, arguments: []expression{
			integerLiteral(left), integerLiteral(right),
		}}
}

func integerLiteral(value int64) expression {
	return expression{kind: "literal", typeID: TypeInteger,
		literal: buildInteger(value)}
}

func booleanLiteral(value bool) expression {
	return expression{kind: "literal", typeID: TypeBoolean,
		literal: buildBoolean(value)}
}
