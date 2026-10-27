package authorization

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	internal "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/authorization"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

var NewDecisionResult = internal.NewDecisionResult
var NewPolicyDecision = internal.NewPolicyDecision
var NewAllVisibility = internal.NewAllVisibility
var NewNoVisibility = internal.NewNoVisibility
var NewPredicateVisibility = internal.NewPredicateVisibility
var NewMandatoryPredicate = internal.NewMandatoryPredicate
var NewMandatoryPredicateFilter = internal.NewMandatoryPredicateFilter
var NewMandatoryAnd = internal.NewMandatoryAnd
var NewMandatoryOr = internal.NewMandatoryOr
var NewMandatoryNot = internal.NewMandatoryNot
var NewScalarValue = resolved.NewScalarValue
var NewBooleanValue = resolved.NewBooleanValue
var NewJSONValue = resolved.NewJSONValue

func NewDeniedResult(id [sha256.Size]byte) (DecisionResult, error) {
	return internal.NewDecisionResult(id, false)
}

func NewEnumValue(value string) (Value, error) {
	return resolved.NewScalarValue(api.LogicalTypeEnum, value)
}
