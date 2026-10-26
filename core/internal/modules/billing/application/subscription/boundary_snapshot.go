package subscriptionapp

import (
	"time"

	"github.com/google/uuid"

	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
)

type BoundarySnapshot struct {
	Kind             BoundaryKind
	DueAt            time.Time
	Change           *SubscriptionChange
	TargetPrice      *TargetPrice
	NextPeriod       *billingperiod.Period
	PaymentReceiptID *uuid.UUID
	GraceEndsAt      *time.Time
}

func (boundary BoundarySnapshot) valid() bool {
	if !boundary.Kind.valid() || !validInstant(boundary.DueAt) ||
		(boundary.Change != nil && !boundary.Change.valid()) ||
		(boundary.TargetPrice != nil && !boundary.TargetPrice.valid()) ||
		(boundary.NextPeriod != nil && (!validInstant(boundary.NextPeriod.Start()) ||
			!validInstant(boundary.NextPeriod.End()))) ||
		(boundary.PaymentReceiptID != nil && !validV7(*boundary.PaymentReceiptID)) ||
		(boundary.GraceEndsAt != nil && !validInstant(*boundary.GraceEndsAt)) {
		return false
	}
	switch boundary.Kind {
	case BoundaryManualChange:
		return boundary.Change != nil && boundary.TargetPrice != nil &&
			(boundary.PaymentReceiptID == nil || boundary.NextPeriod != nil)
	case BoundaryTrialEnd:
		return boundary.Change == nil && boundary.TargetPrice == nil &&
			((boundary.PaymentReceiptID == nil && boundary.NextPeriod == nil) ||
				(boundary.PaymentReceiptID != nil && boundary.NextPeriod != nil))
	case BoundaryPeriodEnd:
		return boundary.Change == nil && boundary.TargetPrice == nil &&
			boundary.GraceEndsAt == nil &&
			((boundary.PaymentReceiptID == nil && boundary.NextPeriod == nil) ||
				(boundary.PaymentReceiptID != nil && boundary.NextPeriod != nil))
	case BoundaryCancellation, BoundaryGraceEnd:
		return boundary.Change == nil && boundary.TargetPrice == nil &&
			boundary.NextPeriod == nil &&
			boundary.PaymentReceiptID == nil && boundary.GraceEndsAt == nil
	default:
		return false
	}
}
