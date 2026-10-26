package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type SearchTerm struct {
	field     api.FieldDefinition
	value     Value
	path      RequestPath
	validated bool
}

func NewSearchTerm(
	field api.FieldDefinition, value Value, path RequestPath,
) (SearchTerm, error) {
	term := SearchTerm{
		field: field, value: cloneValue(value), path: path, validated: true,
	}
	if err := term.Validate(); err != nil {
		return SearchTerm{}, err
	}
	return term, nil
}

func (term SearchTerm) Validate() error {
	if !term.validated || term.field.Validate() != nil || term.value.Validate() != nil ||
		term.path.Validate() != nil {
		return fmt.Errorf("resolved REST Query search term is invalid")
	}
	capabilities, err := term.field.Capabilities()
	if err != nil {
		return fmt.Errorf("resolved REST Query search term is invalid")
	}
	searchable, err := capabilities.Searchable()
	if err != nil || !searchable {
		return fmt.Errorf("resolved REST Query search term is invalid")
	}
	fieldType, err := term.field.Type()
	if err != nil {
		return fmt.Errorf("resolved REST Query search term is invalid")
	}
	if !valueConformsToFieldType(fieldType, api.OperatorContains, term.value) {
		return fmt.Errorf("resolved REST Query search term is invalid")
	}
	return nil
}

func (term SearchTerm) Path() (RequestPath, error) {
	if err := term.Validate(); err != nil {
		return RequestPath{}, err
	}
	return term.path, nil
}

func (term SearchTerm) Field() (api.FieldDefinition, error) {
	if err := term.Validate(); err != nil {
		return api.FieldDefinition{}, err
	}
	return term.field, nil
}

func (term SearchTerm) Value() (Value, error) {
	if err := term.Validate(); err != nil {
		return Value{}, err
	}
	return cloneValue(term.value), nil
}

func cloneSearchTerms(values []SearchTerm) []SearchTerm {
	result := make([]SearchTerm, len(values))
	for index, value := range values {
		result[index] = value
		result[index].value = cloneValue(value.value)
	}
	return result
}

func (SearchTerm) String() string                      { return valueRedaction }
func (SearchTerm) GoString() string                    { return valueRedaction }
func (term SearchTerm) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (SearchTerm) MarshalJSON() ([]byte, error)        { return rejectDirectJSONMarshal() }
