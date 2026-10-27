package catalog

import (
	"math"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

type coreFieldSpec struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Column         string                 `json:"column"`
	Kind           string                 `json:"kind"`
	EnumVersion    string                 `json:"enum_version,omitempty"`
	EnumValues     []string               `json:"enum_values,omitempty"`
	Operators      []string               `json:"operators"`
	Sortable       bool                   `json:"sortable"`
	CursorBytes    int                    `json:"cursor_bytes"`
	Searchable     bool                   `json:"searchable"`
	FoldedColumn   string                 `json:"folded_column,omitempty"`
	Classification profile.Classification `json:"classification"`
	Default        bool                   `json:"default"`
	CreateRequired bool                   `json:"create_required,omitempty"`
	MaximumBytes   int                    `json:"maximum_bytes,omitempty"`
}

func newCoreField(spec coreFieldSpec) (api.FieldDefinition, error) {
	fieldType, err := coreFieldType(spec)
	if err != nil {
		return api.FieldDefinition{}, ErrContract
	}
	column, err := sqlidentifier.MintStatic(spec.Column)
	if err != nil {
		return api.FieldDefinition{}, ErrContract
	}
	var folded *sqlidentifier.Token
	if spec.Searchable != (spec.FoldedColumn != "") {
		return api.FieldDefinition{}, ErrContract
	}
	if spec.FoldedColumn != "" {
		token, tokenErr := sqlidentifier.MintStatic(spec.FoldedColumn)
		if tokenErr != nil {
			return api.FieldDefinition{}, ErrContract
		}
		folded = &token
	}
	storage, err := api.NewFieldStorage(column, folded, nil)
	if err != nil {
		return api.FieldDefinition{}, ErrContract
	}
	capabilities, err := fieldCapabilities(fieldType, spec.Operators,
		spec.Sortable, spec.CursorBytes, spec.Searchable)
	if err != nil {
		return api.FieldDefinition{}, ErrContract
	}
	return api.NewFieldDefinition(spec.ID, spec.Name, fieldType, storage, capabilities)
}

func coreFieldType(spec coreFieldSpec) (api.FieldType, error) {
	switch spec.Kind {
	case "text":
		casefold := api.TextCasefoldDisabled
		if spec.Searchable {
			casefold = api.TextCasefoldUnicodeDefaultV1
		}
		text, err := api.NewTextProfile(casefold, api.TextAccentDisabled)
		if err != nil {
			return api.FieldType{}, err
		}
		return api.NewTextFieldType(text)
	case "integer":
		integer, err := api.NewIntegerProfile(1, math.MaxInt64)
		if err != nil {
			return api.FieldType{}, err
		}
		return api.NewIntegerFieldType(integer)
	case "boolean":
		return api.NewBooleanFieldType()
	case "timestamp":
		return api.NewTimestampFieldType()
	case "uuid":
		return api.NewUUIDFieldType()
	case "relation":
		return api.NewRelationFieldType()
	case "uuid_list":
		return api.NewUUIDListFieldType()
	case "enum":
		return coreEnumType(spec.EnumVersion, spec.EnumValues)
	default:
		return api.FieldType{}, ErrContract
	}
}

func coreEnumType(version string, values []string) (api.FieldType, error) {
	if uuid.Validate(version) != nil || len(values) == 0 {
		return api.FieldType{}, ErrContract
	}
	enumValues := make([]api.EnumValue, len(values))
	for index, code := range values {
		value, err := api.NewEnumValue(code, nil)
		if err != nil {
			return api.FieldType{}, ErrContract
		}
		enumValues[index] = value
	}
	enum, err := api.NewEnumProfile(version, enumValues)
	if err != nil {
		return api.FieldType{}, ErrContract
	}
	return api.NewEnumFieldType(enum)
}
