package recordkernel

import (
	"bytes"
	"testing"
)

func TestDecodeValueSupportsAllPublishedTypes(t *testing.T) {
	uuid := `"71000000-0000-4000-8000-000000000001"`
	cases := []struct {
		name     string
		field    Field
		raw      string
		expected string
	}{
		{"text", field(TypeText), `"e\u0301"`, `"é"`},
		{"integer", field(TypeInteger), `"-42"`, `"-42"`},
		{"integer_number", field(TypeInteger), `-42`, `"-42"`},
		{"decimal", decimalField(), `"12.3"`, `"12.30"`},
		{"decimal_number", decimalField(), `12.3`, `"12.30"`},
		{"boolean", field(TypeBoolean), `true`, `true`},
		{"date", field(TypeDate), `"2026-07-17"`, `"2026-07-17"`},
		{"timestamp", field(TypeTimestamp),
			`"2026-07-17T10:11:12.123456Z"`, `"2026-07-17T10:11:12.123456Z"`},
		{"uuid", field(TypeUUID), uuid, uuid},
		{"enum", enumField(TypeEnum), `"OPEN"`, `"OPEN"`},
		{"relation", field(TypeRelation), uuid, uuid},
		{"text_list", field(TypeTextList), `["e\u0301","x"]`, `["é","x"]`},
		{"uuid_list", field(TypeUUIDList), `[` + uuid + `]`, `[` + uuid + `]`},
		{"json", field(TypeJSON), `{"b":2,"a":1}`, `{"a":1,"b":2}`},
		{"time", field(TypeTime), `"10:11:12.123456"`, `"10:11:12.123456"`},
		{"document_ref", field(TypeDocumentRef), uuid, uuid},
		{"document_ref_list", field(TypeDocumentRefList),
			`[` + uuid + `]`, `[` + uuid + `]`},
		{"enum_list", enumField(TypeEnumList), `["OPEN"]`, `["OPEN"]`},
	}
	if len(cases) != 18 {
		t.Fatalf("codec matrix has %d entries", len(cases))
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			value, err := DecodeValue(test.field, []byte(test.raw))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(value.CanonicalJSON(), []byte(test.expected)) {
				t.Fatalf("canonical = %s", value.CanonicalJSON())
			}
		})
	}
}

func TestTemporalAndJSONBudgetsFailClosed(t *testing.T) {
	if _, err := DecodeValue(field(TypeDate), []byte(`"0000-01-01"`)); err == nil {
		t.Fatal("year zero was accepted")
	}
	deep := bytes.Repeat([]byte("["), MaximumJSONDepth+1)
	deep = append(deep, '0')
	deep = append(deep, bytes.Repeat([]byte("]"), MaximumJSONDepth+1)...)
	if _, err := DecodeValue(field(TypeJSON), deep); err == nil {
		t.Fatal("over-depth JSON was accepted")
	}
}

func field(typeID TypeID) Field {
	return Field{TypeID: typeID, Nullable: true}
}

func decimalField() Field {
	value := field(TypeDecimal)
	value.Decimal = &DecimalProfile{Precision: 10, Scale: 2}
	return value
}

func enumField(typeID TypeID) Field {
	value := field(typeID)
	value.EnumOptions = []string{"OPEN", "SHUT"}
	return value
}
