package resolved

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
	metrics, valid := filter.measureMandatory(0)
	if !valid || metrics.depth > maximumMandatoryDepth ||
		metrics.nodes > maximumMandatoryNodes ||
		metrics.predicates > maximumMandatoryPredicates ||
		metrics.literals > maximumMandatoryLiterals {
		return fmt.Errorf("resolved REST Query mandatory filter is invalid")
	}
	return nil
}

func (filter MandatoryFilter) measureMandatory(
	depth int,
) (mandatoryMetrics, bool) {
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
	minimum := 2
	if filter.kind == api.RequestFilterNot {
		minimum = 1
	} else if filter.kind != api.RequestFilterAnd &&
		filter.kind != api.RequestFilterOr {
		return mandatoryMetrics{}, false
	}
	if filter.predicate != nil || len(filter.children) < minimum ||
		filter.kind == api.RequestFilterNot && len(filter.children) != 1 {
		return mandatoryMetrics{}, false
	}
	metrics.depth = depth + 1
	for _, child := range filter.children {
		childMetrics, valid := child.measureMandatory(depth + 1)
		if !valid {
			return mandatoryMetrics{}, false
		}
		metrics.nodes += childMetrics.nodes
		metrics.predicates += childMetrics.predicates
		metrics.literals += childMetrics.literals
		metrics.depth = max(metrics.depth, childMetrics.depth)
	}
	return metrics, true
}

func (predicate MandatoryPredicate) Validate() error {
	if !predicate.validated || predicate.field.Validate() != nil ||
		predicate.values == nil {
		return fmt.Errorf("resolved REST Query mandatory predicate is invalid")
	}
	fieldType, err := predicate.field.Type()
	if err != nil || !operatorAllowedForType(fieldType, predicate.operator) {
		return fmt.Errorf("resolved REST Query mandatory predicate is invalid")
	}
	arity, valid := predicate.operator.Arity()
	if !valid || !resolvedArityValid(arity, len(predicate.values)) ||
		len(predicate.values) > maximumMandatoryLiterals {
		return fmt.Errorf("resolved REST Query mandatory predicate is invalid")
	}
	for _, value := range predicate.values {
		if !valueConformsToFieldType(fieldType, predicate.operator, value) {
			return fmt.Errorf("resolved REST Query mandatory predicate is invalid")
		}
	}
	return nil
}
