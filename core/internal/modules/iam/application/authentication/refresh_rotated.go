package authentication

import "time"

type RefreshOutcome string

const (
	RefreshRotatedOutcome  RefreshOutcome = "rotated"
	RefreshReusedOutcome   RefreshOutcome = "reused"
	RefreshExpiredOutcome  RefreshOutcome = "expired"
	RefreshRejectedOutcome RefreshOutcome = "rejected"
)

type RefreshRotated struct {
	Outcome  RefreshOutcome
	IssuedAt time.Time
}

func (rotated RefreshRotated) validate() error {
	switch rotated.Outcome {
	case RefreshRotatedOutcome:
		if rotated.IssuedAt.IsZero() {
			return ErrAuthenticationStorage
		}
	case RefreshReusedOutcome, RefreshExpiredOutcome, RefreshRejectedOutcome:
		if !rotated.IssuedAt.IsZero() {
			return ErrAuthenticationStorage
		}
	default:
		return ErrAuthenticationStorage
	}
	return nil
}
