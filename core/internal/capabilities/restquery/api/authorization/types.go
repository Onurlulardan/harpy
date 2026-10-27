package authorization

import (
	internal "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/authorization"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type PolicyPort = internal.PolicyPort
type GlobalPolicyPort = internal.GlobalPolicyPort
type PolicyRegistration = internal.PolicyRegistration
type PolicyRequest = internal.PolicyRequest
type GlobalPolicyRequest = internal.GlobalPolicyRequest
type PolicyDecision = internal.PolicyDecision
type Requirement = internal.Requirement
type RequirementKind = internal.RequirementKind
type DecisionResult = internal.DecisionResult
type Visibility = internal.Visibility
type VisibilityMode = internal.VisibilityMode
type MandatoryPredicate = internal.MandatoryPredicate
type MandatoryFilter = internal.MandatoryFilter
type Value = resolved.Value

const (
	RequirementResourceRead     = internal.RequirementResourceRead
	RequirementFieldSelect      = internal.RequirementFieldSelect
	RequirementFieldFilter      = internal.RequirementFieldFilter
	RequirementFieldSort        = internal.RequirementFieldSort
	RequirementFieldSearch      = internal.RequirementFieldSearch
	RequirementFieldGroup       = internal.RequirementFieldGroup
	RequirementFieldMeasure     = internal.RequirementFieldMeasure
	RequirementFieldHaving      = internal.RequirementFieldHaving
	RequirementRelationInclude  = internal.RequirementRelationInclude
	RequirementExactCount       = internal.RequirementExactCount
	RequirementAggregateMeasure = internal.RequirementAggregateMeasure
	RequirementAggregateHaving  = internal.RequirementAggregateHaving
	RequirementAggregateSort    = internal.RequirementAggregateSort
	VisibilityAll               = internal.VisibilityAll
	VisibilityNone              = internal.VisibilityNone
	VisibilityPredicate         = internal.VisibilityPredicate
)
