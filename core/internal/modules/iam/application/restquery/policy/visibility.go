package policy

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
)

func resourceVisibility(
	requirementID [32]byte,
	resource catalog.Resource,
	allowed bool,
) (restauthorization.Visibility, error) {
	if !allowed {
		visibility, err := restauthorization.NewNoVisibility(requirementID)
		if err != nil {
			return restauthorization.Visibility{}, catalog.ErrContract
		}
		return visibility, nil
	}
	value, err := restauthorization.NewEnumValue(resource.TerminalStatus())
	if err != nil {
		return restauthorization.Visibility{}, catalog.ErrContract
	}
	predicate, err := restauthorization.NewMandatoryPredicate(
		resource.StatusFieldID().String(), api.LogicalTypeEnum,
		api.OperatorNe, []restauthorization.Value{value},
	)
	if err != nil {
		return restauthorization.Visibility{}, catalog.ErrContract
	}
	filter, err := restauthorization.NewMandatoryPredicateFilter(predicate)
	if err != nil {
		return restauthorization.Visibility{}, catalog.ErrContract
	}
	visibility, err := restauthorization.NewPredicateVisibility(requirementID, filter)
	if err != nil {
		return restauthorization.Visibility{}, catalog.ErrContract
	}
	return visibility, nil
}
