package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type AggregateGroup struct {
	field     api.FieldDefinition
	key       OutputKey
	path      RequestPath
	validated bool
}

func NewAggregateGroup(
	field api.FieldDefinition, key OutputKey, path RequestPath,
) (AggregateGroup, error) {
	group := AggregateGroup{field: field, key: key, path: path, validated: true}
	if err := group.Validate(); err != nil {
		return AggregateGroup{}, err
	}
	return group, nil
}

func (group AggregateGroup) Validate() error {
	if !group.validated || group.field.Validate() != nil || group.key.Validate() != nil ||
		group.path.Validate() != nil {
		return fmt.Errorf("resolved REST Query aggregate group is invalid")
	}
	capabilities, err := group.field.Capabilities()
	if err != nil {
		return fmt.Errorf("resolved REST Query aggregate group is invalid")
	}
	allowed, err := capabilities.Groupable()
	if err != nil || !allowed {
		return fmt.Errorf("resolved REST Query aggregate group is invalid")
	}
	return nil
}

func (group AggregateGroup) Field() (api.FieldDefinition, error) {
	if err := group.Validate(); err != nil {
		return api.FieldDefinition{}, err
	}
	return group.field, nil
}

func (group AggregateGroup) Key() (OutputKey, error) {
	if err := group.Validate(); err != nil {
		return OutputKey{}, err
	}
	return group.key, nil
}

func (group AggregateGroup) Path() (RequestPath, error) {
	if err := group.Validate(); err != nil {
		return RequestPath{}, err
	}
	return group.path, nil
}

func (AggregateGroup) String() string                       { return valueRedaction }
func (AggregateGroup) GoString() string                     { return valueRedaction }
func (group AggregateGroup) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (AggregateGroup) MarshalJSON() ([]byte, error)         { return rejectDirectJSONMarshal() }
