package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type Projection struct {
	field     api.FieldDefinition
	key       OutputKey
	path      RequestPath
	hidden    bool
	validated bool
}

func NewProjection(
	field api.FieldDefinition, key OutputKey, path RequestPath, hidden bool,
) (Projection, error) {
	projection := Projection{
		field: field, key: key, path: path, hidden: hidden, validated: true,
	}
	if err := projection.Validate(); err != nil {
		return Projection{}, err
	}
	return projection, nil
}

func (projection Projection) Validate() error {
	if !projection.validated || projection.field.Validate() != nil ||
		projection.key.Validate() != nil || projection.path.Validate() != nil {
		return fmt.Errorf("resolved REST Query projection is invalid")
	}
	capabilities, err := projection.field.Capabilities()
	if err != nil {
		return fmt.Errorf("resolved REST Query projection is invalid")
	}
	selectable, err := capabilities.Selectable()
	if err != nil || !selectable {
		return fmt.Errorf("resolved REST Query projection is invalid")
	}
	return nil
}

func (projection Projection) Path() (RequestPath, error) {
	if err := projection.Validate(); err != nil {
		return RequestPath{}, err
	}
	return projection.path, nil
}

func (projection Projection) Field() (api.FieldDefinition, error) {
	if err := projection.Validate(); err != nil {
		return api.FieldDefinition{}, err
	}
	return projection.field, nil
}

func (projection Projection) Key() (OutputKey, error) {
	if err := projection.Validate(); err != nil {
		return OutputKey{}, err
	}
	return projection.key, nil
}

func (projection Projection) Hidden() (bool, error) {
	if err := projection.Validate(); err != nil {
		return false, err
	}
	return projection.hidden, nil
}

func (Projection) String() string                            { return valueRedaction }
func (Projection) GoString() string                          { return valueRedaction }
func (projection Projection) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Projection) MarshalJSON() ([]byte, error)              { return rejectDirectJSONMarshal() }
