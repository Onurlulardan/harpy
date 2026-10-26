package api

import "errors"

var (
	ErrOnboardingInvalid      = errors.New("tenancy onboarding input is invalid")
	ErrOnboardingConflict     = errors.New("tenancy onboarding operation conflicts")
	ErrOnboardingUnavailable  = errors.New("tenancy onboarding is unavailable")
	ErrOnboardingLeaseHeld    = errors.New("tenancy onboarding lease is held")
	ErrOnboardingLeaseExpired = errors.New("tenancy onboarding lease expired")
	ErrOnboardingReplay       = errors.New("tenancy onboarding replay rejected")
	ErrAuthorityCASConflict   = errors.New("protected authority CAS conflict")
)
