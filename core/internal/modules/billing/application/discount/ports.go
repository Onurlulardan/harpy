package discount

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	Redeem(context.Context, RedeemExecution, RedeemPlanner) (RedeemResult, error)
}

type RedeemPlanner func(EligibilitySnapshot, *Redemption) (RedemptionPlan, error)

type IDGenerator func() (uuid.UUID, error)
type Clock func() time.Time
