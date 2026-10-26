package recordkernel

import (
	"errors"
	"strings"
	"testing"
)

func TestTextCallMatrixAndOutputBudget(t *testing.T) {
	tests := []struct {
		name     string
		function string
		inputs   []string
		expected string
	}{
		{"concat", "concat", []string{"e\u0301", "x"}, "éx"},
		{"lower", "lower", []string{"HELLO"}, "hello"},
		{"upper", "upper", []string{"hello"}, "HELLO"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			arguments := make([]expression, len(test.inputs))
			for index, input := range test.inputs {
				arguments[index] = textLiteral(t, input)
			}
			value, err := evaluateTestExpression(expression{kind: "call",
				typeID: TypeText, function: test.function,
				arguments: arguments}, nil)
			text, _ := valueString(value)
			if err != nil || text != test.expected {
				t.Fatalf("text = %q, err = %v", text, err)
			}
		})
	}
	accepted := strings.Repeat("x", MaximumFieldBytes-2)
	half := len(accepted) / 2
	if _, err := evaluateTestExpression(concatExpression(t,
		accepted[:half], accepted[half:]), nil); err != nil {
		t.Fatalf("1 MiB concat: %v", err)
	}
	if _, err := evaluateTestExpression(concatExpression(t,
		accepted[:half], accepted[half:]+"x"), nil); !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("1 MiB + 1 concat: %v", err)
	}
}

func TestCoalesceShortCircuitsAndNullFallsThrough(t *testing.T) {
	invalid := binaryInteger("divide", 1, 0)
	coalesce := expression{kind: "call", typeID: TypeInteger,
		function: "coalesce", arguments: []expression{integerLiteral(7), invalid}}
	catalog := fixtureCatalog(nil)
	budget := newInstructionBudget(catalog)
	value, err := evaluateExpression(coalesce, Record{}, nil, &budget)
	if err != nil || string(value.CanonicalJSON()) != `"7"` || budget.used != 2 {
		t.Fatalf("coalesce = %s, used = %d, err = %v",
			value.CanonicalJSON(), budget.used, err)
	}
	coalesce.arguments[0].literal = nullValue(TypeInteger)
	if _, err := evaluateTestExpression(coalesce, nil); !errors.Is(err, ErrCalculation) {
		t.Fatalf("fall-through error = %v", err)
	}
}

func TestRoundAndDateAddExactMatrix(t *testing.T) {
	profile := &DecimalProfile{Precision: 8, Scale: 2}
	for _, test := range []struct {
		input    string
		expected string
	}{{"2.345", `"2.34"`}, {"2.355", `"2.36"`}, {"-2.355", `"-2.36"`}} {
		value, err := evaluateTestExpression(roundExpression(t, test.input, 2), profile)
		if err != nil || string(value.CanonicalJSON()) != test.expected {
			t.Fatalf("round %s = %s, %v", test.input, value.CanonicalJSON(), err)
		}
	}
	if _, err := evaluateTestExpression(roundExpression(t, "2.35", 3), profile); !errors.Is(err, ErrCalculation) {
		t.Fatalf("invalid scale = %v", err)
	}
	date := dateAddExpression(t, TypeDate, "2024-02-28", 1)
	value, err := evaluateTestExpression(date, nil)
	if err != nil || string(value.CanonicalJSON()) != `"2024-02-29"` {
		t.Fatalf("date_add = %s, %v", value.CanonicalJSON(), err)
	}
	timestamp := dateAddExpression(t, TypeTimestamp,
		"2026-07-17T10:11:12.123456Z", -1)
	value, err = evaluateTestExpression(timestamp, nil)
	if err != nil || string(value.CanonicalJSON()) !=
		`"2026-07-16T10:11:12.123456Z"` {
		t.Fatalf("timestamp date_add = %s, %v", value.CanonicalJSON(), err)
	}
	if _, err := evaluateTestExpression(
		dateAddExpression(t, TypeDate, "9999-12-31", 1), nil,
	); !errors.Is(err, ErrCalculation) {
		t.Fatalf("date overflow = %v", err)
	}
}

func textLiteral(t *testing.T, value string) expression {
	t.Helper()
	typed, err := buildText(value)
	if err != nil {
		t.Fatal(err)
	}
	return expression{kind: "literal", typeID: TypeText, literal: typed}
}

func concatExpression(t *testing.T, values ...string) expression {
	arguments := make([]expression, len(values))
	for index, value := range values {
		arguments[index] = textLiteral(t, value)
	}
	return expression{kind: "call", typeID: TypeText,
		function: "concat", arguments: arguments}
}

func dateAddExpression(
	t *testing.T, typeID TypeID, value string, days int64,
) expression {
	t.Helper()
	typed, err := decodeTemporal(typeID, []byte(`"`+value+`"`))
	if err != nil {
		t.Fatal(err)
	}
	return expression{kind: "call", typeID: typeID, function: "date_add",
		arguments: []expression{{kind: "literal", typeID: typeID, literal: typed},
			integerLiteral(days)}}
}
