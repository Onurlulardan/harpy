package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type AggregateSort struct {
	reference AggregateReference
	direction api.Direction
	implicit  bool
	validated bool
}

func NewAggregateSort(
	reference AggregateReference, direction api.Direction, implicit bool,
) (AggregateSort, error) {
	sort := AggregateSort{
		reference: reference, direction: direction,
		implicit: implicit, validated: true,
	}
	if err := sort.Validate(); err != nil {
		return AggregateSort{}, err
	}
	return sort, nil
}

func (sort AggregateSort) Validate() error {
	if !sort.validated || sort.reference.Validate() != nil || sort.direction.String() == "" {
		return fmt.Errorf("resolved REST Query aggregate sort is invalid")
	}
	return nil
}

func (sort AggregateSort) Reference() (AggregateReference, error) {
	if err := sort.Validate(); err != nil {
		return AggregateReference{}, err
	}
	return sort.reference, nil
}

func (sort AggregateSort) Direction() (api.Direction, error) {
	if err := sort.Validate(); err != nil {
		return 0, err
	}
	return sort.direction, nil
}

func (sort AggregateSort) Implicit() (bool, error) {
	if err := sort.Validate(); err != nil {
		return false, err
	}
	return sort.implicit, nil
}

func (AggregateSort) String() string                      { return valueRedaction }
func (AggregateSort) GoString() string                    { return valueRedaction }
func (sort AggregateSort) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (AggregateSort) MarshalJSON() ([]byte, error)        { return rejectDirectJSONMarshal() }
