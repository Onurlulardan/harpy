package resolver

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func FuzzResolveLiteralsNeverPanics(f *testing.F) {
	f.Add(uint8(0), []byte(`"text"`))
	f.Add(uint8(1), []byte("42"))
	f.Add(uint8(2), []byte(`{"a":1}`))
	f.Add(uint8(3), []byte{0xff, 0x00})
	fields := resolverFuzzFieldTypes()
	f.Fuzz(func(t *testing.T, control uint8, data []byte) {
		literal, valid := fuzzLiteral(control, data)
		if !valid {
			return
		}
		fieldType := fields[int(control)%len(fields)]
		values, err := ResolveLiterals(fieldType, api.OperatorEq, []api.Literal{literal})
		if err != nil {
			var resolutionError LiteralResolutionError
			if !errors.As(err, &resolutionError) {
				t.Fatalf("resolution error type = %T", err)
			}
			if _, codeErr := resolutionError.Code(); codeErr != nil {
				t.Fatalf("resolution code error = %v", codeErr)
			}
			return
		}
		for _, value := range values {
			if err := value.Validate(); err != nil {
				t.Fatalf("resolved value is invalid: %v", err)
			}
		}
	})
}

func fuzzLiteral(control uint8, data []byte) (api.Literal, bool) {
	if control&1 == 0 {
		literal, err := api.NewJSONLiteral(data)
		return literal, err == nil
	}
	literal, err := api.NewTextLiteral(string(data))
	return literal, err == nil
}

func resolverFuzzFieldTypes() []api.FieldType {
	textProfile, _ := api.NewTextProfile(api.TextCasefoldDisabled, api.TextAccentDisabled)
	textType, _ := api.NewTextFieldType(textProfile)
	integerProfile, _ := api.NewIntegerProfile(-100, 100)
	integerType, _ := api.NewIntegerFieldType(integerProfile)
	decimalProfile, _ := api.NewDecimalProfile(8, 2)
	decimalType, _ := api.NewDecimalFieldType(decimalProfile)
	booleanType, _ := api.NewBooleanFieldType()
	dateType, _ := api.NewDateFieldType()
	uuidType, _ := api.NewUUIDFieldType()
	jsonType, _ := api.NewJSONFieldType()
	return []api.FieldType{
		textType, integerType, decimalType, booleanType, dateType, uuidType, jsonType,
	}
}
