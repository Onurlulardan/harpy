package storagecatalog

import (
	"math"

	restapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	workquery "github.com/Onurlulardan/harpy/core/internal/modules/work/application/restquery"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func buildField(resource string, spec queryFieldSpec) (restapi.FieldDefinition, error) {
	fieldType, err := fieldType(spec.kind)
	if err != nil {
		return restapi.FieldDefinition{}, workquery.ErrUnavailable
	}
	column, err := sqlidentifier.MintStatic(spec.column)
	if err != nil {
		return restapi.FieldDefinition{}, workquery.ErrUnavailable
	}
	var folded *sqlidentifier.Token
	if spec.folded != "" {
		value, foldedErr := sqlidentifier.MintStatic(spec.folded)
		if foldedErr != nil {
			return restapi.FieldDefinition{}, workquery.ErrUnavailable
		}
		folded = &value
	}
	storage, err := restapi.NewFieldStorage(column, folded, nil)
	if err != nil {
		return restapi.FieldDefinition{}, workquery.ErrUnavailable
	}
	filters, err := filters(spec)
	if err != nil {
		return restapi.FieldDefinition{}, workquery.ErrUnavailable
	}
	cursorBytes := 0
	if spec.sortable {
		cursorBytes = cursorBytesFor(spec.kind)
	}
	capabilities, err := restapi.NewFieldCapabilities(fieldType,
		restapi.FieldCapabilitiesInput{Selectable: spec.selectable,
			Filters: filters, Sortable: spec.sortable,
			CursorCanonicalBytes: cursorBytes, Searchable: spec.searchable})
	if err != nil {
		return restapi.FieldDefinition{}, workquery.ErrUnavailable
	}
	return restapi.NewFieldDefinition(queryID(resource, "field:"+spec.name).String(),
		spec.name, fieldType, storage, capabilities)
}

func fieldType(kind string) (restapi.FieldType, error) {
	switch kind {
	case "uuid":
		return restapi.NewUUIDFieldType()
	case "boolean":
		return restapi.NewBooleanFieldType()
	case "timestamp":
		return restapi.NewTimestampFieldType()
	case "date":
		return restapi.NewDateFieldType()
	case "integer":
		profile, err := restapi.NewIntegerProfile(0, math.MaxInt64)
		if err != nil {
			return restapi.FieldType{}, err
		}
		return restapi.NewIntegerFieldType(profile)
	case "text", "search_text", "text_list":
		casefold := restapi.TextCasefoldDisabled
		if kind == "search_text" {
			casefold = restapi.TextCasefoldUnicodeDefaultV1
		}
		profile, err := restapi.NewTextProfile(casefold, restapi.TextAccentDisabled)
		if err != nil {
			return restapi.FieldType{}, err
		}
		if kind == "text_list" {
			return restapi.NewTextListFieldType(profile)
		}
		return restapi.NewTextFieldType(profile)
	default:
		return restapi.FieldType{}, workquery.ErrUnavailable
	}
}

func cursorBytesFor(kind string) int {
	switch kind {
	case "uuid":
		return 36
	case "boolean":
		return 5
	case "timestamp":
		return 27
	case "date":
		return 10
	case "integer":
		return 20
	default:
		return 512
	}
}
