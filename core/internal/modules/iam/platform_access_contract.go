package iam

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPlatformAccessAuthentication = errors.New("IAM platform access authentication failed")
	ErrPlatformAccessForbidden      = errors.New("IAM platform access is forbidden")
	ErrPlatformAccessUnavailable    = errors.New("IAM platform access is unavailable")
)

type PlatformAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Capability          string
	ExpectedMode        string
	RequireStepUp       bool
	StepUpPurpose       [32]byte
}

type PlatformAccessDecision struct {
	SubjectID, TenantID, SessionID     uuid.UUID
	DeploymentMode                     string
	Capability                         string
	StepUpAt                           time.Time
	PlatformClock                      int64
	TokenVersion, AuthorizationVersion int64
}

func (decision PlatformAccessDecision) Valid() bool {
	return decision.SubjectID != uuid.Nil && decision.TenantID != uuid.Nil &&
		decision.SessionID != uuid.Nil &&
		(decision.DeploymentMode == "onpremise" || decision.DeploymentMode == "saas") &&
		decision.Capability != "" && decision.PlatformClock > 0 &&
		decision.TokenVersion > 0 && decision.AuthorizationVersion > 0
}
