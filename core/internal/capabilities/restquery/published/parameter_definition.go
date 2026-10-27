package published

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type ParameterInput struct {
	Name     string
	Type     api.FieldType
	Arity    ParameterArity
	Required bool
	Default  []resolved.Value
}

type ParameterDefinition struct {
	name      string
	fieldType api.FieldType
	arity     ParameterArity
	required  bool
	defaults  []resolved.Value
	validated bool
}

func (definition ParameterDefinition) Name() (string, error) {
	if err := definition.Validate(); err != nil {
		return "", err
	}
	return definition.name, nil
}

func (definition ParameterDefinition) Type() (api.FieldType, error) {
	if err := definition.Validate(); err != nil {
		return api.FieldType{}, err
	}
	return cloneParameterFieldType(definition.fieldType)
}

func (definition ParameterDefinition) Arity() (ParameterArity, error) {
	if err := definition.Validate(); err != nil {
		return 0, err
	}
	return definition.arity, nil
}

func (definition ParameterDefinition) Required() (bool, error) {
	if err := definition.Validate(); err != nil {
		return false, err
	}
	return definition.required, nil
}

func (definition ParameterDefinition) Default() ([]resolved.Value, bool, error) {
	if err := definition.Validate(); err != nil {
		return nil, false, err
	}
	if definition.defaults == nil {
		return nil, false, nil
	}
	return cloneResolvedValues(definition.defaults), true, nil
}

func (definition ParameterDefinition) String() string   { return publishedRedaction }
func (definition ParameterDefinition) GoString() string { return publishedRedaction }
func (definition ParameterDefinition) Format(state fmt.State, _ rune) {
	writePublishedRedaction(state)
}
func (ParameterDefinition) MarshalJSON() ([]byte, error) { return rejectPublishedJSON() }
func (ParameterInput) String() string                    { return publishedRedaction }
func (ParameterInput) GoString() string                  { return publishedRedaction }
func (ParameterInput) Format(state fmt.State, _ rune)    { writePublishedRedaction(state) }
func (ParameterInput) MarshalJSON() ([]byte, error)      { return rejectPublishedJSON() }

func (definition ParameterDefinition) clone() ParameterDefinition {
	result := definition
	result.fieldType, _ = cloneParameterFieldType(definition.fieldType)
	result.defaults = cloneResolvedValues(definition.defaults)
	return result
}
