package recordkernel

import (
	"errors"
	"testing"
)

func TestEvaluatorAccepts100000AndRejects100001Instructions(t *testing.T) {
	catalog := instructionCatalog()
	input, err := DecodeInput(catalog, ActionCreate,
		fixtureActorAuthority(), []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteMutation(catalog, Record{}, input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Instructions() != MaximumInstructions {
		t.Fatalf("instructions = %d", result.Instructions())
	}
	catalog.crossFieldValidations = []crossFieldRule{{
		id: fixtureV7(999), code: "LIMIT_CHECK",
		predicate: expression{kind: "literal", typeID: TypeBoolean,
			literal: buildBoolean(true)},
	}}
	if _, err := ExecuteMutation(catalog, Record{}, input); !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("100001 instructions: %v", err)
	}
}

func TestRoundHalfEvenAndNullPropagation(t *testing.T) {
	catalog := fixtureCatalog(nil)
	budget := newInstructionBudget(catalog)
	round := roundExpression(t, "2.345", 2)
	profile := &DecimalProfile{Precision: 8, Scale: 2}
	value, err := evaluateExpression(round, Record{}, profile, &budget)
	if err != nil {
		t.Fatal(err)
	}
	coerced, err := coerceExpressionValue(Field{TypeID: TypeDecimal,
		Decimal: &DecimalProfile{Precision: 8, Scale: 2}}, value)
	if err != nil || string(coerced.CanonicalJSON()) != `"2.34"` {
		t.Fatalf("half-even down = %s, %v", coerced.CanonicalJSON(), err)
	}
	strict := expression{kind: "binary", typeID: TypeInteger,
		operator: "add", arguments: []expression{
			{kind: "literal", typeID: TypeInteger,
				literal: nullValue(TypeInteger)},
			{kind: "literal", typeID: TypeInteger,
				literal: buildInteger(1)},
		}}
	value, err = evaluateExpression(strict, Record{}, nil, &budget)
	if err != nil || !value.IsNull() {
		t.Fatalf("strict null = %+v, %v", value, err)
	}
}

func TestDecimalIntermediateOverflowCannotBeCancelled(t *testing.T) {
	left := decimalLiteral(t, "999.99")
	increment := decimalLiteral(t, "0.01")
	decrement := decimalLiteral(t, "1.00")
	add := expression{kind: "binary", typeID: TypeDecimal,
		operator: "add", arguments: []expression{left, increment}}
	root := expression{kind: "binary", typeID: TypeDecimal,
		operator: "subtract", arguments: []expression{add, decrement}}
	catalog := fixtureCatalog(nil)
	budget := newInstructionBudget(catalog)
	profile := &DecimalProfile{Precision: 5, Scale: 2}
	if _, err := evaluateExpression(root, Record{}, profile, &budget); err == nil {
		t.Fatal("intermediate precision overflow was cancelled")
	}
}

func TestStringComparisonUsesDecodedNFCScalarValues(t *testing.T) {
	newline, _ := decodeText(TypeText, []byte(`"a\nb"`))
	quote, _ := decodeText(TypeText, []byte(`"a\"b"`))
	comparison, err := compareValues(newline, quote)
	if err != nil || comparison >= 0 {
		t.Fatalf("decoded ordering = %d, %v", comparison, err)
	}
	decomposed, _ := decodeText(TypeText, []byte(`"e\u0301"`))
	composed, _ := decodeText(TypeText, []byte(`"é"`))
	comparison, err = compareValues(decomposed, composed)
	if err != nil || comparison != 0 {
		t.Fatalf("NFC comparison = %d, %v", comparison, err)
	}
}

func instructionCatalog() Catalog {
	fields := make([]Field, 100)
	calculations := make([]calculation, len(fields))
	for index := range fields {
		fields[index] = Field{ID: fixtureV7(index + 100),
			PublicName: "calculation", TypeID: TypeInteger,
			Role: RoleCalculation, Nullable: true,
			WritePolicy: WriteEngine, Generation: 1}
		calculations[index] = calculation{fields[index].ID, unaryChain(1000)}
	}
	catalog := fixtureCatalog(fields)
	catalog.calculations = calculations
	return catalog
}

func unaryChain(nodes int) expression {
	result := expression{kind: "literal", typeID: TypeInteger,
		literal: buildInteger(1)}
	for index := 1; index < nodes; index++ {
		result = expression{kind: "unary", typeID: TypeInteger,
			operator: "negate", arguments: []expression{result}}
	}
	return result
}

func roundExpression(t *testing.T, decimal string, scale int64) expression {
	t.Helper()
	value, err := decodeGenericDecimal([]byte(`"` + decimal + `"`))
	if err != nil {
		t.Fatal(err)
	}
	return expression{kind: "call", typeID: TypeDecimal,
		function: "round", arguments: []expression{
			{kind: "literal", typeID: TypeDecimal, literal: value},
			{kind: "literal", typeID: TypeInteger,
				literal: buildInteger(scale)},
		}}
}

func decimalLiteral(t *testing.T, decimal string) expression {
	t.Helper()
	value, err := decodeGenericDecimal([]byte(`"` + decimal + `"`))
	if err != nil {
		t.Fatal(err)
	}
	return expression{kind: "literal", typeID: TypeDecimal, literal: value}
}
