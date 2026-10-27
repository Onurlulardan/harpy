package rowdecode

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestRowScannerUsesExactPGXDestinationsAndOwnsFieldTypes(t *testing.T) {
	jsonType, _ := api.NewJSONFieldType()
	fieldTypes := []api.FieldType{textType(t), jsonType}
	scanner, err := NewRowScanner(fieldTypes)
	if err != nil {
		t.Fatal(err)
	}
	fieldTypes[0] = api.FieldType{}
	destinations, err := scanner.Destinations()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := destinations[0].(*any); !ok {
		t.Fatalf("text destination = %T", destinations[0])
	}
	if _, ok := destinations[1].(*[]byte); !ok {
		t.Fatalf("JSON destination = %T", destinations[1])
	}
	pgxScanValue(t, pgtype.TextOID, "owned", destinations[0])
	pgxScanValue(t, pgtype.JSONBOID, json.RawMessage(`{"b":2,"a":1}`), destinations[1])
	values, err := scanner.Decode()
	if err != nil || len(values) != 2 {
		t.Fatalf("decode scanned row: %v", err)
	}
	requireResponse(t, values[0], `"owned"`)
	requireResponse(t, values[1], `{"a":1,"b":2}`)
}

func TestRowScannerFailsClosedOnPartialTamperedAndReusedFlow(t *testing.T) {
	jsonType, _ := api.NewJSONFieldType()
	fieldTypes := []api.FieldType{textType(t), jsonType}

	partial, _ := NewRowScanner(fieldTypes)
	destinations, _ := partial.Destinations()
	pgxScanValue(t, pgtype.TextOID, "only-one", destinations[0])
	if values, err := partial.Decode(); err == nil || values != nil {
		t.Fatal("partial scan returned a row")
	}
	if values, err := partial.Decode(); err == nil || values != nil {
		t.Fatal("failed scanner was reusable")
	}

	tampered, _ := NewRowScanner(fieldTypes[:1])
	destinations, _ = tampered.Destinations()
	destinations[0] = new(any)
	if values, err := tampered.Decode(); err == nil || values != nil {
		t.Fatal("replaced destination returned a row")
	}

	reused, _ := NewRowScanner(fieldTypes[:1])
	destinations, _ = reused.Destinations()
	pgxScanValue(t, pgtype.TextOID, "once", destinations[0])
	if _, err := reused.Decode(); err != nil {
		t.Fatal(err)
	}
	if destinations, err := reused.Destinations(); err == nil || destinations != nil {
		t.Fatal("consumed scanner issued destinations")
	}

	premature, _ := NewRowScanner(fieldTypes[:1])
	if values, err := premature.Decode(); err == nil || values != nil {
		t.Fatal("decode before destination issuance succeeded")
	}
	if destinations, err := premature.Destinations(); err == nil || destinations != nil {
		t.Fatal("prematurely consumed scanner was reusable")
	}

	doubleIssue, _ := NewRowScanner(fieldTypes[:1])
	if _, err := doubleIssue.Destinations(); err != nil {
		t.Fatal(err)
	}
	if destinations, err := doubleIssue.Destinations(); err == nil || destinations != nil {
		t.Fatal("destinations issued twice")
	}
	if values, err := doubleIssue.Decode(); err == nil || values != nil {
		t.Fatal("double-issued scanner remained usable")
	}
}

func TestRowScannerDistinguishesSQLNullFromJSONLiteralNull(t *testing.T) {
	jsonType, _ := api.NewJSONFieldType()
	storedNull, _ := NewRowScanner([]api.FieldType{jsonType})
	destinations, _ := storedNull.Destinations()
	pgxScanValue(t, pgtype.JSONBOID, json.RawMessage(`null`), destinations[0])
	values, err := storedNull.Decode()
	if err != nil {
		t.Fatal(err)
	}
	isNull, _ := values[0].IsNull()
	if isNull {
		t.Fatal("stored JSON null became SQL NULL")
	}

	sqlNull, _ := NewRowScanner([]api.FieldType{jsonType})
	destinations, _ = sqlNull.Destinations()
	if err := pgtype.NewMap().Scan(
		pgtype.JSONBOID, pgtype.BinaryFormatCode, nil, destinations[0],
	); err != nil {
		t.Fatal(err)
	}
	values, err = sqlNull.Decode()
	if err != nil {
		t.Fatal(err)
	}
	isNull, _ = values[0].IsNull()
	if !isNull {
		t.Fatal("SQL NULL identity was lost")
	}
}

func TestRowScannerRejectsInvalidConstructionAndRedactsScannedData(t *testing.T) {
	if scanner, err := NewRowScanner(nil); err == nil || scanner != nil {
		t.Fatal("empty scanner accepted")
	}
	if scanner, err := NewRowScanner([]api.FieldType{{}}); err == nil || scanner != nil {
		t.Fatal("invalid field type accepted")
	}
	tooMany := make([]api.FieldType, maximumRowColumns+1)
	for index := range tooMany {
		tooMany[index] = textType(t)
	}
	if scanner, err := NewRowScanner(tooMany); err == nil || scanner != nil {
		t.Fatal("oversized scanner accepted")
	}

	scanner, _ := NewRowScanner([]api.FieldType{textType(t)})
	destinations, _ := scanner.Destinations()
	pgxScanValue(t, pgtype.TextOID, "top-secret", destinations[0])
	for _, rendered := range []string{
		fmt.Sprint(scanner), fmt.Sprintf("%#v", scanner),
		scanner.LogValue().String(), slog.Any("scanner", scanner).Value.String(),
	} {
		if rendered != redacted || strings.Contains(rendered, "top-secret") {
			t.Fatalf("scanner leaked: %q", rendered)
		}
	}
	if encoded, err := json.Marshal(scanner); err == nil || len(encoded) != 0 {
		t.Fatal("scanner marshaled directly")
	}
}
