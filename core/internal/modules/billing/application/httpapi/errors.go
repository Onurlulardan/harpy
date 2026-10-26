package httpapi

import "errors"

var (
	ErrInvalid              = errors.New("billing HTTP input is invalid")
	ErrAuthentication       = errors.New("billing authentication failed")
	ErrForbidden            = errors.New("billing operation is forbidden")
	ErrNotFound             = errors.New("billing resource was not found")
	ErrConflict             = errors.New("billing operation conflicts")
	ErrRevisionStale        = errors.New("billing root revision is stale")
	ErrImmutableConflict    = errors.New("billing immutable resource conflicts")
	ErrTransitionInvalid    = errors.New("billing transition is invalid")
	ErrIdempotencyReused    = errors.New("billing idempotency key was reused")
	ErrIdempotencyProgress  = errors.New("billing idempotency is in progress")
	ErrAuthorityUnavailable = errors.New("billing authority is unavailable")
	ErrUnavailable          = errors.New("billing operation is unavailable")
)

type UsageLimitExceeded struct {
	EntitlementKey string
	Current        int64
	Requested      int64
	Limit          int64
	PeriodStart    string
	PeriodEnd      string
	ResetPolicy    string
}

func (usage UsageLimitExceeded) Error() string { return "USAGE_LIMIT_EXCEEDED" }

func (usage UsageLimitExceeded) Valid() bool {
	return usage.EntitlementKey != "" && usage.Current >= 0 && usage.Requested > 0 &&
		usage.Limit >= 0 && usage.Current <= usage.Limit &&
		usage.Requested > usage.Limit-usage.Current &&
		usage.PeriodStart != "" && usage.PeriodEnd != "" &&
		(usage.ResetPolicy == "billing_period" ||
			usage.ResetPolicy == "calendar_month_utc" || usage.ResetPolicy == "never")
}
