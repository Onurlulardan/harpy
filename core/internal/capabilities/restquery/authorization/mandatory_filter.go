package authorization

import (
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

type MandatoryFilter struct {
	kind      api.RequestFilterKind
	predicate *MandatoryPredicate
	children  []MandatoryFilter
	validated bool
}

type MandatoryPredicate struct {
	fieldID     uuid.UUID
	logicalType api.LogicalType
	operator    api.Operator
	values      []resolved.Value
	validated   bool
}

func NewMandatoryPredicate(
	fieldID string,
	logicalType api.LogicalType,
	operator api.Operator,
	values []resolved.Value,
) (MandatoryPredicate, error) {
	identifier, err := uuid.Parse(fieldID)
	predicate := MandatoryPredicate{
		fieldID: identifier, logicalType: logicalType, operator: operator,
		values: append([]resolved.Value{}, values...),
		validated: err == nil && identifier != uuid.Nil && identifier.String() == fieldID &&
			identifier.Variant() == uuid.RFC4122,
	}
	if err := predicate.Validate(); err != nil {
		return MandatoryPredicate{}, err
	}
	return predicate, nil
}

func NewMandatoryPredicateFilter(
	predicate MandatoryPredicate,
) (MandatoryFilter, error) {
	owned := predicate.clone()
	filter := MandatoryFilter{
		kind: api.RequestFilterPredicate, predicate: &owned, validated: true,
	}
	if err := filter.Validate(); err != nil {
		return MandatoryFilter{}, err
	}
	return filter, nil
}

func NewMandatoryAnd(children []MandatoryFilter) (MandatoryFilter, error) {
	return newMandatoryBoolean(api.RequestFilterAnd, children)
}

func NewMandatoryOr(children []MandatoryFilter) (MandatoryFilter, error) {
	return newMandatoryBoolean(api.RequestFilterOr, children)
}

func NewMandatoryNot(child MandatoryFilter) (MandatoryFilter, error) {
	return newMandatoryBoolean(api.RequestFilterNot, []MandatoryFilter{child})
}

func newMandatoryBoolean(
	kind api.RequestFilterKind,
	children []MandatoryFilter,
) (MandatoryFilter, error) {
	filter := MandatoryFilter{
		kind: kind, children: cloneMandatoryFilters(children), validated: true,
	}
	if err := filter.Validate(); err != nil {
		return MandatoryFilter{}, err
	}
	return filter, nil
}

func (MandatoryFilter) String() string   { return authorizationRedaction }
func (MandatoryFilter) GoString() string { return authorizationRedaction }
func (MandatoryFilter) LogValue() slog.Value {
	return authorizationLogValue()
}
func (filter MandatoryFilter) Format(state fmt.State, _ rune) {
	writeAuthorizationRedaction(state)
}
func (MandatoryFilter) MarshalJSON() ([]byte, error) { return rejectAuthorizationJSON() }

func (MandatoryPredicate) String() string   { return authorizationRedaction }
func (MandatoryPredicate) GoString() string { return authorizationRedaction }
func (MandatoryPredicate) LogValue() slog.Value {
	return authorizationLogValue()
}
func (predicate MandatoryPredicate) Format(state fmt.State, _ rune) {
	writeAuthorizationRedaction(state)
}
func (MandatoryPredicate) MarshalJSON() ([]byte, error) {
	return rejectAuthorizationJSON()
}
