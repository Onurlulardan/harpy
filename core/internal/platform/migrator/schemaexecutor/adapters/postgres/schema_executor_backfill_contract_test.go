package postgres

import (
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func TestSchemaExecutorConverterTypes(t *testing.T) {
	tests := []struct {
		name       string
		converter  schemav1.Converter
		source     string
		target     string
		compatible bool
	}{
		{"identity", schemav1.ConvertIdentity, "text", "text", true},
		{"identity mismatch", schemav1.ConvertIdentity, "bigint", "text", false},
		{"integer decimal", schemav1.ConvertIntegerDecimal, "bigint", "numeric(18,2)", true},
		{"wrong source", schemav1.ConvertIntegerDecimal, "text", "numeric(18,2)", false},
		{"integer text", schemav1.ConvertIntegerText, "bigint", "text", true},
		{"decimal text", schemav1.ConvertDecimalText, "numeric(18,2)", "text", true},
		{"date timestamp", schemav1.ConvertDateTimestamp, "date", "timestamp with time zone", true},
		{"uuid text", schemav1.ConvertUUIDText, "uuid", "text", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := schemaExecutorConverterTypes(test.converter, test.source, test.target)
			if actual != test.compatible {
				t.Fatalf("compatibility = %t, want %t", actual, test.compatible)
			}
		})
	}
}
