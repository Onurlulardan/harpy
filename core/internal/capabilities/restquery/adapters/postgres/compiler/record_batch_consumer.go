package compiler

import (
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type RecordBatchConsumer interface {
	ConsumeRecord(RecordStatement) error
	ConsumeExactCount(ExactCountStatement) error
}

// ConsumeRecordBatch is the only plan-bound leaf-consumption boundary.
func ConsumeRecordBatch(
	batch RecordBatch,
	plan resolved.AuthorizedPlan,
	consumer RecordBatchConsumer,
) error {
	if batch.ValidatePlan(plan) != nil {
		return ErrContract
	}
	return consumeValidatedRecordBatch(batch, consumer)
}

func consumeValidatedRecordBatch(
	batch RecordBatch, consumer RecordBatchConsumer,
) error {
	if batch.Validate() != nil || recordBatchConsumerIsNil(consumer) {
		return ErrContract
	}
	if err := consumer.ConsumeRecord(batch.record); err != nil {
		return err
	}
	if batch.hasExactCount {
		if err := consumer.ConsumeExactCount(batch.exactCount); err != nil {
			return err
		}
	}
	return nil
}

func recordBatchConsumerIsNil(consumer RecordBatchConsumer) bool {
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
