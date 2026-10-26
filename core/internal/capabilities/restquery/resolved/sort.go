package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type Sort struct {
	field     api.FieldDefinition
	direction api.Direction
	path      RequestPath
	implicit  bool
	validated bool
}

func NewSort(
	field api.FieldDefinition,
	direction api.Direction,
	path RequestPath,
	implicit bool,
) (Sort, error) {
	sort := Sort{
		field: field, direction: direction, path: path,
		implicit: implicit, validated: true,
	}
	if err := sort.Validate(); err != nil {
		return Sort{}, err
	}
	return sort, nil
}

func (sort Sort) Implicit() (bool, error) {
	if err := sort.Validate(); err != nil {
		return false, err
	}
	return sort.implicit, nil
}

func (sort Sort) Validate() error {
	if !sort.validated || sort.field.Validate() != nil || sort.direction.String() == "" ||
		sort.path.Validate() != nil {
		return fmt.Errorf("resolved REST Query sort is invalid")
	}
	capabilities, err := sort.field.Capabilities()
	if err != nil {
		return fmt.Errorf("resolved REST Query sort is invalid")
	}
	sortable, err := capabilities.Sortable()
	bound, boundErr := capabilities.CursorCanonicalBytes()
	if err != nil || boundErr != nil || !sortable || bound < 1 || bound > 2048 {
		return fmt.Errorf("resolved REST Query sort is invalid")
	}
	return nil
}

func (sort Sort) Path() (RequestPath, error) {
	if err := sort.Validate(); err != nil {
		return RequestPath{}, err
	}
	return sort.path, nil
}

func (sort Sort) Field() (api.FieldDefinition, error) {
	if err := sort.Validate(); err != nil {
		return api.FieldDefinition{}, err
	}
	return sort.field, nil
}

func (sort Sort) Direction() (api.Direction, error) {
	if err := sort.Validate(); err != nil {
		return 0, err
	}
	return sort.direction, nil
}

func (Sort) String() string                      { return valueRedaction }
func (Sort) GoString() string                    { return valueRedaction }
func (sort Sort) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Sort) MarshalJSON() ([]byte, error)        { return rejectDirectJSONMarshal() }
