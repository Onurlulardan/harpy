package restquery

import (
	"crypto/sha256"
	"strconv"

	restapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	"github.com/google/uuid"
)

func workQueryVisibility(resourceID [sha256.Size]byte, catalog ResolvedCatalog,
	evidence Evidence, principals []string, revision int64, allowed bool,
) (restauthorization.Visibility, error) {
	if !allowed {
		value, err := restauthorization.NewNoVisibility(resourceID)
		if err != nil {
			return restauthorization.Visibility{}, ErrHTTPUnavailable
		}
		return value, nil
	}
	filters := make([]restauthorization.MandatoryFilter, 0, 4)
	visibility, err := principalVisibility(catalog, "visibility_principals", principals)
	if err != nil {
		return restauthorization.Visibility{}, err
	}
	filters = append(filters, visibility)
	if evidence.resource == "inbox" || evidence.resource == "calendar-unified" {
		contextFilter, filterErr := principalVisibility(catalog, "context_principals", principals)
		freshFilter, freshErr := projectionFreshVisibility(catalog, revision)
		if filterErr != nil || freshErr != nil {
			return restauthorization.Visibility{}, ErrHTTPUnavailable
		}
		filters = append(filters, contextFilter, freshFilter)
	}
	if evidence.resource == "comments" {
		parent, err := uuid.Parse(evidence.ids["task_id"])
		if err != nil || parent == uuid.Nil {
			return restauthorization.Visibility{}, ErrHTTPInvalid
		}
		parentFilter, err := scalarVisibility(catalog, "task_id", restapi.LogicalTypeUUID,
			restapi.OperatorEq, parent.String())
		if err != nil {
			return restauthorization.Visibility{}, err
		}
		filters = append(filters, parentFilter)
	}
	combined, err := restauthorization.NewMandatoryAnd(filters)
	if err != nil {
		return restauthorization.Visibility{}, ErrHTTPUnavailable
	}
	result, err := restauthorization.NewPredicateVisibility(resourceID, combined)
	if err != nil {
		return restauthorization.Visibility{}, ErrHTTPUnavailable
	}
	return result, nil
}

func principalVisibility(catalog ResolvedCatalog, field string,
	principals []string,
) (restauthorization.MandatoryFilter, error) {
	values := make([]restauthorization.Value, len(principals))
	for index, principal := range principals {
		value, err := restauthorization.NewScalarValue(restapi.LogicalTypeText, principal)
		if err != nil {
			return restauthorization.MandatoryFilter{}, ErrHTTPUnavailable
		}
		values[index] = value
	}
	fieldID, err := catalog.fieldID(field)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrHTTPUnavailable
	}
	predicate, err := restauthorization.NewMandatoryPredicate(
		fieldID.String(),
		restapi.LogicalTypeTextList, restapi.OperatorContainsAny, values)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrHTTPUnavailable
	}
	return restauthorization.NewMandatoryPredicateFilter(predicate)
}

func projectionFreshVisibility(catalog ResolvedCatalog, revision int64) (
	restauthorization.MandatoryFilter, error,
) {
	owner, err := scalarVisibility(catalog, "source_owner", restapi.LogicalTypeText,
		restapi.OperatorEq, "WORK")
	fresh, freshErr := scalarVisibility(catalog, "permission_revision",
		restapi.LogicalTypeInteger, restapi.OperatorEq, strconv.FormatInt(revision, 10))
	if err != nil || freshErr != nil {
		return restauthorization.MandatoryFilter{}, ErrHTTPUnavailable
	}
	return restauthorization.NewMandatoryOr([]restauthorization.MandatoryFilter{owner, fresh})
}

func scalarVisibility(catalog ResolvedCatalog, field string, kind restapi.LogicalType,
	operator restapi.Operator, raw string,
) (restauthorization.MandatoryFilter, error) {
	value, err := restauthorization.NewScalarValue(kind, raw)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrHTTPUnavailable
	}
	fieldID, err := catalog.fieldID(field)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrHTTPUnavailable
	}
	predicate, err := restauthorization.NewMandatoryPredicate(
		fieldID.String(), kind, operator,
		[]restauthorization.Value{value})
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrHTTPUnavailable
	}
	return restauthorization.NewMandatoryPredicateFilter(predicate)
}
