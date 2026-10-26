package compiler

import (
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type AggregateBatchConsumer interface {
	ConsumeAggregate(AggregateStatement) error
}

// ConsumeAggregateBatch is the only plan-bound aggregate leaf boundary.
func ConsumeAggregateBatch(
	batch AggregateBatch,
	plan resolved.AuthorizedPlan,
	consumer AggregateBatchConsumer,
) error {
	if batch.ValidatePlan(plan) != nil || aggregateConsumerNil(consumer) {
		return ErrContract
	}
	return consumer.ConsumeAggregate(batch.statement)
}

func aggregateConsumerNil(consumer AggregateBatchConsumer) bool {
	if consumer == nil {
		return true
	}
	value := reflect.ValueOf(consumer)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
