package subscriptionapp

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

type UpdateFromProviderRequest struct {
	SubscriptionID    uuid.UUID
	ProviderReceiptID uuid.UUID
	NormalizedFactID  uuid.UUID
	NormalizedFactSHA [32]byte
	TargetState       subscriptiondomain.State
	OccurredAt        time.Time
	Period            *billingperiod.Period
	TrialEndsAt       *time.Time
	GraceEndsAt       *time.Time
}

func (commands *Commands) UpdateFromProvider(
	ctx context.Context,
	request UpdateFromProviderRequest,
) (Result, error) {
	now, err := commands.instant()
	if err != nil || ctx == nil || ctx.Err() != nil || !providerRequestValid(request) ||
		request.OccurredAt.After(now) {
		return Result{}, ErrRequest
	}
	digest := providerDigest(request)
	execution := ProviderExecution{
		Execution: Execution{
			OperationID:       OperationUpdateFromProvider,
			SubscriptionID:    request.SubscriptionID,
			SourceOperationID: request.ProviderReceiptID, RequestSHA256: digest,
		},
		ProviderReceiptID: request.ProviderReceiptID,
		NormalizedFactID:  request.NormalizedFactID,
		NormalizedFactSHA: request.NormalizedFactSHA,
	}
	result, err := commands.store.Provider(
		ctx, execution, func(locked LockedState) (CommitPlan, error) {
			if !locked.valid() || locked.Subscription.ID != request.SubscriptionID {
				return CommitPlan{}, ErrSnapshot
			}
			if replay, found, replayErr := receiptReplay(
				locked.Receipt, execution.Execution,
			); found {
				if replayErr != nil {
					return CommitPlan{}, ErrProviderFactReused
				}
				return replay, nil
			}
			if subscriptiondomain.ValidateTransition(
				locked.Subscription.State, request.TargetState,
			) != nil {
				return CommitPlan{}, ErrStateConflict
			}
			after, applyErr := providerSnapshot(locked.Subscription, request, now)
			if applyErr != nil {
				return CommitPlan{}, applyErr
			}
			identifiers, identifierErr := commands.baseIdentifiers()
			if identifierErr != nil {
				return CommitPlan{}, identifierErr
			}
			plan := basePlan(planInputs{
				Before: locked.Subscription, After: after, Execution: execution.Execution,
				ReceiptID: identifiers.receipt, AuditID: identifiers.audit,
				EventID:            identifiers.subscriptionEvent,
				EntitlementEventID: identifiers.entitlementEvent, Now: now,
				EventActorID: execution.SourceOperationID,
				Kind:         "provider_state_transition", Reason: "verified_normalized_provider_fact",
			})
			plan.CancelJobs = boundaryJobCancellations(
				locked, now, "provider_state_transition",
			)
			if kind, runAt, scheduled := nextBoundary(after); scheduled {
				job, jobErr := commands.boundaryJob(after.ID, kind, runAt, after.Revision)
				if jobErr != nil {
					return CommitPlan{}, jobErr
				}
				plan.Jobs = []JobIntent{job}
			}
			if !plan.valid() {
				return CommitPlan{}, ErrConfiguration
			}
			return plan, nil
		},
	)
	if err != nil {
		return Result{}, fmt.Errorf("update billing subscription from provider: %w", err)
	}
	return checkedResult(result, request.SubscriptionID)
}

func providerRequestValid(request UpdateFromProviderRequest) bool {
	return validID(request.SubscriptionID) && validV7(request.ProviderReceiptID) &&
		validV7(request.NormalizedFactID) && validDigest(request.NormalizedFactSHA) &&
		request.TargetState.Valid() &&
		validInstant(request.OccurredAt) && providerWindowValid(request)
}

func providerWindowValid(request UpdateFromProviderRequest) bool {
	if request.Period != nil && (!validInstant(request.Period.Start()) ||
		!validInstant(request.Period.End())) ||
		request.TrialEndsAt != nil && !validInstant(*request.TrialEndsAt) ||
		request.GraceEndsAt != nil && !validInstant(*request.GraceEndsAt) {
		return false
	}
	switch request.TargetState {
	case subscriptiondomain.Trialing:
		return request.Period != nil && request.TrialEndsAt != nil &&
			request.TrialEndsAt.Equal(request.Period.End()) && request.GraceEndsAt == nil
	case subscriptiondomain.Active:
		return request.Period != nil && request.TrialEndsAt == nil && request.GraceEndsAt == nil
	case subscriptiondomain.PastDue:
		return request.GraceEndsAt != nil && request.TrialEndsAt == nil
	default:
		return request.TrialEndsAt == nil && request.GraceEndsAt == nil
	}
}

func providerSnapshot(
	before SubscriptionSnapshot,
	request UpdateFromProviderRequest,
	now time.Time,
) (SubscriptionSnapshot, error) {
	after := before.revised(now)
	after.State, after.Period = request.TargetState, request.Period
	after.TrialEndsAt = copiedInstant(request.TrialEndsAt)
	after.GraceEndsAt = copiedInstant(request.GraceEndsAt)
	if after.State.Terminal() || after.Period == nil || after.CancelAt == nil ||
		!after.CancelAt.Equal(after.Period.End()) {
		after.CancelAtPeriodEnd, after.CancelAt = false, nil
	}
	if !after.valid() {
		return SubscriptionSnapshot{}, ErrSnapshot
	}
	return after, nil
}

func providerDigest(request UpdateFromProviderRequest) [32]byte {
	writer := newDigestWriter()
	writer.text(OperationUpdateFromProvider)
	writer.identifier(request.SubscriptionID)
	writer.identifier(request.ProviderReceiptID)
	writer.identifier(request.NormalizedFactID)
	writer.digest(request.NormalizedFactSHA)
	writer.text(string(request.TargetState))
	writer.instant(request.OccurredAt)
	if request.Period == nil {
		writer.bytes([]byte{0})
	} else {
		writer.bytes([]byte{1})
		writer.instant(request.Period.Start())
		writer.instant(request.Period.End())
	}
	writer.optionalInstant(request.TrialEndsAt)
	writer.optionalInstant(request.GraceEndsAt)
	return writer.sum()
}
