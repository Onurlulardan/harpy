package recordpage

import (
	"encoding/json"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func TestEncodeRecordPreservesDeclarationOrderAndHidesSortColumn(t *testing.T) {
	textType := testTextType(t)
	uuidType, _ := api.NewUUIDFieldType()
	name, _ := rowdecode.Decode(textType, "Ada")
	role, _ := rowdecode.Decode(textType, "admin")
	hiddenID := uuid.MustParse("30000000-0000-4000-8000-000000000007")
	hidden, _ := rowdecode.Decode(uuidType, [16]byte(hiddenID))
	columns := []columnMetadata{
		{fieldType: textType, kind: api.LogicalTypeText,
			outputKey: "name", hasOutput: true},
		{fieldType: textType, kind: api.LogicalTypeText,
			outputKey: "role", hasOutput: true},
		{fieldType: uuidType, kind: api.LogicalTypeUUID,
			sortOrdinal: 0, sorted: true},
	}
	document, err := encodeRecord([]rowdecode.Value{name, role, hidden}, columns)
	if err != nil {
		t.Fatal(err)
	}
	if string(document) != `{"name":"Ada","role":"admin"}` {
		t.Fatalf("record=%s", document)
	}
	if json.Valid(document) == false {
		t.Fatal("record is not valid JSON")
	}
}

func TestValidateRowRejectsWidthAndLogicalTypeMismatch(t *testing.T) {
	textType := testTextType(t)
	uuidType, _ := api.NewUUIDFieldType()
	textValue, _ := rowdecode.Decode(textType, "Ada")
	identifier := uuid.MustParse("30000000-0000-4000-8000-000000000001")
	uuidValue, _ := rowdecode.Decode(uuidType, [16]byte(identifier))
	columns := []columnMetadata{{fieldType: textType, kind: api.LogicalTypeText}}
	if validateRow(nil, columns) != ErrContract {
		t.Fatal("short row was accepted")
	}
	if validateRow([]rowdecode.Value{textValue, textValue}, columns) != ErrContract {
		t.Fatal("wide row was accepted")
	}
	if validateRow([]rowdecode.Value{uuidValue}, columns) != ErrContract {
		t.Fatal("wrong logical type was accepted")
	}
}

func FuzzEncodeRecordEscapesOutputKeyAndValue(f *testing.F) {
	f.Add("display_name", "Ada")
	f.Add("quote\"key", "line\nvalue")
	textType := testTextType(f)
	f.Fuzz(func(t *testing.T, key string, raw string) {
		if !validOutputKey(key) {
			t.Skip()
		}
		value, err := rowdecode.Decode(textType, raw)
		if err != nil {
			t.Skip()
		}
		column := columnMetadata{
			fieldID:   uuid.MustParse("20000000-0000-4000-8000-000000000001"),
			fieldType: textType, kind: api.LogicalTypeText,
			outputKey: key, hasOutput: true, sortOrdinal: -1,
		}
		document, err := encodeRecord([]rowdecode.Value{value}, []columnMetadata{column})
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]string
		if json.Unmarshal(document, &decoded) != nil || len(decoded) != 1 || decoded[key] != raw {
			t.Fatalf("document=%q", document)
		}
	})
}

func testTextType(t testing.TB) api.FieldType {
	t.Helper()
	profile, err := api.NewTextProfile(api.TextCasefoldDisabled, api.TextAccentDisabled)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := api.NewTextFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}
