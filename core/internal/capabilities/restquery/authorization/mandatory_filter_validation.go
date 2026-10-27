package authorization

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

const (
	maximumMandatoryDepth      = 5
	maximumMandatoryNodes      = 128
	maximumMandatoryPredicates = 20
	maximumMandatoryLiterals   = 100
)

type mandatoryMetrics struct {
	depth      int
	nodes      int
	predicates int
	literals   int
}

func (filter MandatoryFilter) Validate() error {
	metrics, valid := filter.measure(0)
	if !valid || metrics.depth > maximumMandatoryDepth ||
		metrics.nodes > maximumMandatoryNodes ||
		metrics.predicates > maximumMandatoryPredicates ||
		metrics.literals > maximumMandatoryLiterals {
		return fmt.Errorf("REST Query mandatory filter is invalid")
	}
	return nil
}

func (filter MandatoryFilter) measure(depth int) (mandatoryMetrics, bool) {
	if !filter.validated {
		return mandatoryMetrics{}, false
	}
	metrics := mandatoryMetrics{depth: depth, nodes: 1}
	if filter.kind == api.RequestFilterPredicate {
		if filter.predicate == nil || filter.children != nil ||
			filter.predicate.Validate() != nil {
			return mandatoryMetrics{}, false
		}
		metrics.predicates = 1
		metrics.literals = len(filter.predicate.values)
		return metrics, true
	}
	minimumChildren := 2
	if filter.kind == api.RequestFilterNot {
		minimumChildren = 1
	} else if filter.kind != api.RequestFilterAnd && filter.kind != api.RequestFilterOr {
		return mandatoryMetrics{}, false
	}
	if filter.predicate != nil || len(filter.children) < minimumChildren ||
		filter.kind == api.RequestFilterNot && len(filter.children) != 1 {
		return mandatoryMetrics{}, false
	}
	metrics.depth = depth + 1
	for _, child := range filter.children {
		childMetrics, valid := child.measure(depth + 1)
		if !valid {
			return mandatoryMetrics{}, false
		}
		metrics.nodes += childMetrics.nodes
		metrics.predicates += childMetrics.predicates
		metrics.literals += childMetrics.literals
		if childMetrics.depth > metrics.depth {
			metrics.depth = childMetrics.depth
		}
	}
	return metrics, true
}

func (predicate MandatoryPredicate) Validate() error {
	if !predicate.validated || !validAuthorizationUUID(predicate.fieldID) ||
		!predicate.logicalType.Valid() || predicate.values == nil ||
		!mandatoryArityValid(predicate.operator, len(predicate.values)) ||
		!operatorPossibleForLogicalType(predicate.logicalType, predicate.operator) {
		return fmt.Errorf("REST Query mandatory predicate is invalid")
	}
	expected := mandatoryValueKind(predicate.logicalType)
	for _, value := range predicate.values {
		actual, err := value.Kind()
		if err != nil || actual != expected {
			return fmt.Errorf("REST Query mandatory predicate is invalid")
		}
	}
	return nil
}
