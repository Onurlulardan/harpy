package jobexecution

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

var ErrEventGapProcessorInvalid = errors.New("event gap processor is invalid")

type EventGapReconciliationStore interface {
	ReconcileEventGap(
		context.Context,
		asyncpostgres.JobClaim,
	) (asyncpostgres.EventGapReconciliation, error)
}

type EventGapProcessor struct {
	policy jobsv1.KindPolicy
	store  EventGapReconciliationStore
}

func NewEventGapProcessor(
	policy jobsv1.KindPolicy,
	store EventGapReconciliationStore,
) (*EventGapProcessor, error) {
	canonical, err := platformasync.EventGapJobPolicy(policy.OwnerDomain)
	if err != nil || canonical != policy || nilEventGapStore(store) {
		return nil, ErrEventGapProcessorInvalid
	}
	return &EventGapProcessor{policy: policy, store: store}, nil
}

func (processor *EventGapProcessor) Policy() jobsv1.KindPolicy {
	if processor == nil {
		return jobsv1.KindPolicy{}
	}
	return processor.policy
}

func (processor *EventGapProcessor) Process(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
) error {
	if ctx == nil || processor == nil || nilEventGapStore(processor.store) {
		return ErrEventGapProcessorInvalid
	}
	if _, err := processor.store.ReconcileEventGap(ctx, claim); err != nil {
		return fmt.Errorf("reconcile event sequence gap: %w", err)
	}
	return nil
}

func nilEventGapStore(store EventGapReconciliationStore) bool {
	if store == nil {
		return true
	}
	value := reflect.ValueOf(store)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
