package invoice

import (
	"context"
	"errors"
	"fmt"
	"time"

	invoicedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/invoice"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

func (finalizer *Finalizer) Finalize(
	ctx context.Context,
	request FinalizeRequest,
) (FinalizeResult, error) {
	now := finalizer.clock()
	if !validInstant(now) || !validV7(request.InvoiceID) || !validID(request.ActorID) ||
		!validV7(request.SourceOperationID) || request.ExpectedRevision <= 0 {
		return FinalizeResult{}, ErrRequest
	}
	result, err := finalizer.store.Finalize(
		ctx, FinalizeExecution{Request: request, Now: now},
		func(snapshot DraftSnapshot, existing *Finalization) (FinalizationPlan, error) {
			return finalizer.planFinalization(request, now, snapshot, existing)
		},
	)
	if err != nil {
		if errors.Is(err, ErrSnapshot) || errors.Is(err, ErrState) ||
			errors.Is(err, ErrOperationReused) || errors.Is(err, ErrConfiguration) {
			return FinalizeResult{}, err
		}
		return FinalizeResult{}, fmt.Errorf("finalize billing invoice: %w", err)
	}
	return result, nil
}

func (finalizer *Finalizer) planFinalization(
	request FinalizeRequest,
	now time.Time,
	snapshot DraftSnapshot,
	existing *Finalization,
) (FinalizationPlan, error) {
	if !snapshot.validFor(request) {
		return FinalizationPlan{}, ErrSnapshot
	}
	if existing != nil {
		result, err := finalizationReplay(*existing, request, snapshot)
		return FinalizationPlan{Result: result}, err
	}
	if snapshot.Finalized || snapshot.Revision != request.ExpectedRevision {
		return FinalizationPlan{}, ErrState
	}
	totals, err := invoicedomain.CalculateTotals(snapshot.Lines, snapshot.Settled)
	if err != nil {
		return FinalizationPlan{}, ErrSnapshot
	}
	finalization := Finalization{
		InvoiceID: request.InvoiceID, SourceOperationID: request.SourceOperationID,
		ExpectedRevision: request.ExpectedRevision,
		ResultRevision:   request.ExpectedRevision + 1,
		Totals:           totals, PolicySHA256: snapshot.PolicySHA256, FinalizedAt: now,
	}
	eventID, err := finalizer.nextID()
	if err != nil || !validV7(eventID) {
		return FinalizationPlan{}, ErrConfiguration
	}
	event := FinalizedEventIntent{
		ID: eventID, InvoiceID: request.InvoiceID,
		InvoiceRevision: finalization.ResultRevision,
		ActorID:         request.ActorID, CausationID: request.SourceOperationID,
		OccurredAt: now,
	}
	return FinalizationPlan{
		Result: FinalizeResult{Finalization: finalization}, Write: &finalization,
		Event: &event,
	}, nil
}

func finalizationReplay(
	value Finalization,
	request FinalizeRequest,
	snapshot DraftSnapshot,
) (FinalizeResult, error) {
	expected, err := invoicedomain.CalculateTotals(snapshot.Lines, snapshot.Settled)
	if value.InvoiceID != request.InvoiceID ||
		value.SourceOperationID != request.SourceOperationID ||
		value.ExpectedRevision != request.ExpectedRevision ||
		value.ResultRevision != request.ExpectedRevision+1 ||
		value.PolicySHA256 != snapshot.PolicySHA256 || !validInstant(value.FinalizedAt) ||
		err != nil || !sameTotals(value.Totals, expected) {
		return FinalizeResult{}, ErrOperationReused
	}
	return FinalizeResult{Finalization: value, Replay: true}, nil
}

func sameTotals(left, right invoicedomain.Totals) bool {
	return sameMoney(left.InvoiceTotal(), right.InvoiceTotal()) &&
		sameMoney(left.Settled(), right.Settled()) &&
		sameMoney(left.AmountDue(), right.AmountDue()) &&
		sameMoney(left.CreditBalance(), right.CreditBalance())
}

func sameMoney(left, right money.Money) bool {
	return left.MinorUnits() == right.MinorUnits() && left.Currency() == right.Currency() &&
		left.CurrencyRegistryVersion() == right.CurrencyRegistryVersion()
}
