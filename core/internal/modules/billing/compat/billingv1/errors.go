package billingv1

import "errors"

var (
	ErrRequestInvalid       = errors.New("Billing V1 request is invalid")
	ErrSnapshotRequired     = errors.New("Billing V1 immutable snapshot is required")
	ErrPlanChangeTarget     = errors.New("Billing V1 plan target cannot be resolved")
	ErrUnsupported          = errors.New("Billing V1 mutation is unsupported")
	ErrRetirementEvidence   = errors.New("Billing V1 retirement evidence is unavailable")
	ErrPaginationEvidence   = errors.New("Billing V1 pagination evidence is invalid")
	ErrAvailabilityEvidence = errors.New("Billing V1 availability evidence is invalid")
	ErrRouteUnavailable     = errors.New("Billing V1 route is unavailable in this deployment")
)
