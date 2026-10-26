package deletion

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (worker *Worker) confirm(ctx context.Context, runner JobRunner,
	tenantID domain.ID, task Task, receipt DeleteReceipt,
) (Result, error) {
	billingReceiptID, err := worker.ids.NewUUIDv7()
	if err != nil || !billingReceiptID.UUIDv7() {
		return Result{}, ErrUnavailable
	}
	var result Result
	err = runner.InRepeatableReadJobTenant(ctx, tenantID, func(scoped context.Context,
		scope platformtx.JobTenantScope,
	) error {
		store, bindErr := worker.binder.BindJob(scoped, scope)
		if bindErr != nil || store == nil {
			return ErrUnavailable
		}
		if bindErr = store.BeginReconcile(scoped, task); bindErr != nil {
			return bindErr
		}
		switch task.BillingAction {
		case BillingDecrease:
			bindErr = worker.billing.Decrease(scoped, scope, task.ID,
				task.CiphertextBytes, billingReceiptID)
		case BillingRelease:
			bindErr = worker.billing.Release(scoped, scope,
				task.BillingReservationID, billingReceiptID)
		case BillingKeep:
			bindErr = worker.billing.Keep(scoped, scope, task.ID, billingReceiptID)
		default:
			return ErrUnavailable
		}
		if bindErr != nil {
			return bindErr
		}
		result, bindErr = store.Confirm(scoped, task, receipt, billingReceiptID,
			worker.clock.Now())
		return bindErr
	})
	if err == nil {
		result.DeleteReceipt = receipt
	}
	return result, mapError(err)
}
