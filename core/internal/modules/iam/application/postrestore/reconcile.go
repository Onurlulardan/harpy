package postrestore

import (
	"context"
	"reflect"
)

type Reconciler struct {
	store Store
}

func NewReconciler(store Store) (*Reconciler, error) {
	if nilPostRestoreStore(store) {
		return nil, ErrUnavailable
	}
	return &Reconciler{store: store}, nil
}

func (reconciler *Reconciler) Reconcile(
	ctx context.Context,
	command ReconcileCommand,
) (ReconcileResult, error) {
	if reconciler == nil || nilPostRestoreStore(reconciler.store) ||
		ctx == nil || ctx.Err() != nil || !command.valid() {
		return ReconcileResult{}, ErrDenied
	}
	result, err := reconciler.store.Reconcile(ctx, command)
	if err != nil {
		return ReconcileResult{}, mapStoreError(err)
	}
	if result.State != StateReactivationRequired ||
		len(result.InvalidationCounts) == 0 ||
		result.InvalidationSHA256 == [32]byte{} {
		return ReconcileResult{}, ErrUnavailable
	}
	return result, nil
}

func nilPostRestoreStore(store Store) bool {
	if store == nil {
		return true
	}
	value := reflect.ValueOf(store)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
