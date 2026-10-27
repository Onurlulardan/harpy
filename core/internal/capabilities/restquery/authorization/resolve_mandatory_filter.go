package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveMandatoryFilter(
	catalog api.FieldCatalog,
	filter MandatoryFilter,
) (resolved.MandatoryFilter, error) {
	if catalog.Validate() != nil || filter.Validate() != nil {
		return resolved.MandatoryFilter{}, errPolicyEvaluationContract
	}
	if filter.predicate != nil {
		return resolveMandatoryPredicate(catalog, *filter.predicate)
	}
	children := make([]resolved.MandatoryFilter, 0, len(filter.children))
	for _, child := range filter.children {
		resolvedChild, err := resolveMandatoryFilter(catalog, child)
		if err != nil {
			return resolved.MandatoryFilter{}, err
		}
		children = append(children, resolvedChild)
	}
	switch filter.kind {
	case api.RequestFilterAnd:
		return resolved.NewMandatoryAnd(children)
	case api.RequestFilterOr:
		return resolved.NewMandatoryOr(children)
	case api.RequestFilterNot:
		if len(children) != 1 {
			return resolved.MandatoryFilter{}, errPolicyEvaluationContract
		}
		return resolved.NewMandatoryNot(children[0])
	default:
		return resolved.MandatoryFilter{}, errPolicyEvaluationContract
	}
}

func resolveMandatoryPredicate(
	catalog api.FieldCatalog,
	predicate MandatoryPredicate,
) (resolved.MandatoryFilter, error) {
	field, found, err := catalog.Field(predicate.fieldID)
	if err != nil || !found {
		return resolved.MandatoryFilter{}, errPolicyEvaluationContract
	}
	fieldType, err := field.Type()
	if err != nil {
		return resolved.MandatoryFilter{}, errPolicyEvaluationContract
	}
	kind, err := fieldType.Kind()
	if err != nil || kind != predicate.logicalType {
		return resolved.MandatoryFilter{}, errPolicyEvaluationContract
	}
	resolvedPredicate, err := resolved.NewMandatoryPredicate(
		field, predicate.operator, predicate.values,
	)
	if err != nil {
		return resolved.MandatoryFilter{}, errPolicyEvaluationContract
	}
	filter, err := resolved.NewMandatoryPredicateFilter(resolvedPredicate)
	if err != nil {
		return resolved.MandatoryFilter{}, errPolicyEvaluationContract
	}
	return filter, nil
}
