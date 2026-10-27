package provider

import (
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
	"github.com/google/uuid"
)

const subscriptionSnapshotSchema = "billing.subscription.snapshot"
const subscriptionSnapshotRevision = 1

type subscriptionFactDocument struct {
	Boundary       *boundaryFactDocument `json:"boundary"`
	GraceEndsAt    *string               `json:"grace_ends_at"`
	PeriodEnd      *string               `json:"period_end"`
	PeriodStart    *string               `json:"period_start"`
	State          string                `json:"state"`
	SubscriptionID string                `json:"subscription_id"`
	TrialEndsAt    *string               `json:"trial_ends_at"`
}

type subscriptionFact struct {
	subscriptionID uuid.UUID
	state          subscriptiondomain.State
	period         *billingperiod.Period
	trialEndsAt    *time.Time
	graceEndsAt    *time.Time
	boundary       *boundaryFact
}

func (fact subscriptionFact) request(
	receiptID uuid.UUID,
	eventID uuid.UUID,
	digest [32]byte,
	occurredAt time.Time,
) subscriptionapp.UpdateFromProviderRequest {
	return subscriptionapp.UpdateFromProviderRequest{
		SubscriptionID: fact.subscriptionID, ProviderReceiptID: receiptID,
		NormalizedFactID: eventID, NormalizedFactSHA: digest,
		TargetState: fact.state, OccurredAt: occurredAt, Period: fact.period,
		TrialEndsAt: fact.trialEndsAt, GraceEndsAt: fact.graceEndsAt,
	}
}
