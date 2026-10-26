package invoice

import (
	"context"
	"fmt"
	"time"

	invoiceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/invoice"
)

func (storage *store) Finalize(
	ctx context.Context,
	execution invoiceapp.FinalizeExecution,
	planner invoiceapp.FinalizePlanner,
) (invoiceapp.FinalizeResult, error) {
	if !storage.usable(ctx) || planner == nil || execution.Now.IsZero() ||
		execution.Now.Location() != time.UTC || execution.Now.Nanosecond()%1_000 != 0 {
		return invoiceapp.FinalizeResult{}, invoiceapp.ErrConfiguration
	}
	snapshot, err := storage.lockDraft(ctx, execution.Request.InvoiceID)
	if err != nil {
		return invoiceapp.FinalizeResult{}, fmt.Errorf("lock draft: %w", err)
	}
	existing, found, err := storage.findFinalization(
		ctx, execution.Request.SourceOperationID,
	)
	if err != nil {
		return invoiceapp.FinalizeResult{}, fmt.Errorf("find finalization: %w", err)
	}
	var existingPointer *invoiceapp.Finalization
	if found {
		existingPointer = &existing
	}
	plan, err := planner(snapshot, existingPointer)
	if err != nil {
		return invoiceapp.FinalizeResult{}, err
	}
	if plan.Write == nil {
		if !plan.Result.Replay || existingPointer == nil {
			return invoiceapp.FinalizeResult{}, invoiceapp.ErrSnapshot
		}
		return plan.Result, nil
	}
	if plan.Result.Replay || plan.Result.Finalization != *plan.Write ||
		plan.Event == nil || !plan.Event.Valid() ||
		plan.Event.InvoiceID != plan.Write.InvoiceID ||
		plan.Event.InvoiceRevision != plan.Write.ResultRevision ||
		!plan.Event.OccurredAt.Equal(plan.Write.FinalizedAt) {
		return invoiceapp.FinalizeResult{}, invoiceapp.ErrSnapshot
	}
	if err := storage.writeFinalization(ctx, *plan.Write); err != nil {
		return invoiceapp.FinalizeResult{}, fmt.Errorf("write finalization: %w", err)
	}
	if err := storage.effects.PublishFinalized(ctx, *plan.Event, *plan.Write); err != nil {
		return invoiceapp.FinalizeResult{}, fmt.Errorf("publish finalization: %w", err)
	}
	return plan.Result, nil
}
