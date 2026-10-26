package resolved

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

// ExactCountAuthorized proves that the authorized plan retains exact-count
// intent and that every mandatory root predicate is catalog count-safe.
func (plan AuthorizedPlan) ExactCountAuthorized() (bool, error) {
	if plan.Validate() != nil {
		return false, planValidationError()
	}
	spec, err := plan.Spec()
	if err != nil {
		return false, err
	}
	mode, err := spec.Mode()
	if err != nil || mode != api.QueryModeRecord {
		return false, err
	}
	record, found, err := spec.Record()
	if err != nil || !found {
		return false, err
	}
	count, err := record.Count()
	if err != nil || count != api.CountModeExact {
		return false, err
	}
	reference, referenceErr := spec.Reference()
	rootPath, pathErr := spec.RootPath()
	if referenceErr != nil || pathErr != nil {
		return false, planValidationError()
	}
	rootValue, err := rootPath.Value()
	if err != nil {
		return false, planValidationError()
	}
	for _, visibility := range plan.visibilities {
		catalog, catalogErr := visibility.Catalog()
		path, pathErr := visibility.Path()
		value, valueErr := path.Value()
		if catalogErr != nil || pathErr != nil || valueErr != nil {
			return false, planValidationError()
		}
		if catalog == reference && value == rootValue {
			return visibilityCountSafe(visibility), nil
		}
	}
	return false, planValidationError()
}

func visibilityCountSafe(visibility ResourceVisibility) bool {
	mode, err := visibility.Mode()
	if err != nil {
		return false
	}
	if mode == ResourceVisibilityAll || mode == ResourceVisibilityNone {
		return true
	}
	filter, present, err := visibility.Filter()
	return err == nil && present && mandatoryCountSafe(filter)
}

func mandatoryCountSafe(filter MandatoryFilter) bool {
	predicate, present, err := filter.Predicate()
	if err != nil {
		return false
	}
	if present {
		field, fieldErr := predicate.Field()
		capabilities, capabilityErr := field.Capabilities()
		operator, operatorErr := predicate.Operator()
		capability, found, filterErr := capabilities.Filter(operator)
		countSafe, countErr := capability.CountSafe()
		return fieldErr == nil && capabilityErr == nil && operatorErr == nil &&
			filterErr == nil && found && countErr == nil && countSafe
	}
	children, err := filter.Children()
	if err != nil || len(children) == 0 {
		return false
	}
	for _, child := range children {
		if !mandatoryCountSafe(child) {
			return false
		}
	}
	return true
}

func planValidationError() error {
	return errors.New("resolved REST Query authorized plan is invalid")
}
