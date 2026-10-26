package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type AggregateReference struct {
	kind      api.AggregateReference
	ordinal   int
	path      RequestPath
	validated bool
}

func NewAggregateReference(
	kind api.AggregateReference, ordinal int, path RequestPath,
) (AggregateReference, error) {
	reference := AggregateReference{
		kind: kind, ordinal: ordinal, path: path, validated: true,
	}
	if err := reference.Validate(); err != nil {
		return AggregateReference{}, err
	}
	return reference, nil
}

func (reference AggregateReference) Validate() error {
	if !reference.validated || reference.kind.String() == "" ||
		reference.ordinal < 0 || reference.path.Validate() != nil {
		return fmt.Errorf("resolved REST Query aggregate reference is invalid")
	}
	return nil
}

func (reference AggregateReference) Kind() (api.AggregateReference, error) {
	if err := reference.Validate(); err != nil {
		return 0, err
	}
	return reference.kind, nil
}

func (reference AggregateReference) Ordinal() (int, error) {
	if err := reference.Validate(); err != nil {
		return 0, err
	}
	return reference.ordinal, nil
}

func (reference AggregateReference) Path() (RequestPath, error) {
	if err := reference.Validate(); err != nil {
		return RequestPath{}, err
	}
	return reference.path, nil
}

func (AggregateReference) String() string   { return valueRedaction }
func (AggregateReference) GoString() string { return valueRedaction }
func (reference AggregateReference) Format(state fmt.State, _ rune) {
	writeRedaction(state)
}
func (AggregateReference) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}
