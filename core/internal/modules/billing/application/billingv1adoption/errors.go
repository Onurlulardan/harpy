package billingv1adoptionapp

import "errors"

var (
	ErrPlan     = errors.New("V1 billing plan snapshot is invalid")
	ErrLimits   = errors.New("V1 billing plan limits are invalid")
	ErrFeatures = errors.New("V1 billing plan features are invalid")
)
