package billingv1

import (
	"time"

	"github.com/google/uuid"
)

func NewProduction(dependencies Dependencies) (*Handler, error) {
	dependencies.Clock = time.Now
	dependencies.NextID = uuid.NewV7
	dependencies.CouponCode = secureCouponCode
	return New(dependencies)
}
