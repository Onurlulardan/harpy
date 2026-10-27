package api

import "testing"

func FuzzRequestPredicateNeverPanicsOrAliases(f *testing.F) {
	f.Add("field", uint8(OperatorEq), uint8(1))
	f.Add(string([]byte{0xff}), uint8(OperatorIn), uint8(101))
	f.Fuzz(func(t *testing.T, field string, rawOperator, rawCount uint8) {
		literal, err := NewJSONLiteral([]byte(`"value"`))
		if err != nil {
			t.Fatal(err)
		}
		values := make([]Literal, int(rawCount))
		for index := range values {
			values[index] = literal
		}
		filter, err := NewRequestPredicate(field, Operator(rawOperator), values)
		if err != nil {
			return
		}
		for index := 0; index < 8; index++ {
			next, nextErr := NewRequestNot(filter)
			if nextErr != nil {
				return
			}
			filter = next
		}
	})
}
