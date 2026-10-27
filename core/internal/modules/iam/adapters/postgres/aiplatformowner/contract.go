package aiplatformowner

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrRejected    = errors.New("IAM AI platform owner evidence is rejected")
	ErrUnavailable = errors.New("IAM AI platform owner evidence is unavailable")
)

type Input struct {
	TenantID, ActorID, SessionID                       uuid.UUID
	ExpectedTokenVersion, ExpectedAuthorizationVersion int64
	ExpectedPlatformClock                              int64
}

func (input Input) valid() bool {
	return input.TenantID != uuid.Nil && input.ActorID != uuid.Nil &&
		input.SessionID != uuid.Nil && input.ExpectedTokenVersion > 0 &&
		input.ExpectedAuthorizationVersion > 0 && input.ExpectedPlatformClock > 0
}

type Evidence struct {
	UserID, SessionID, PortalBindingID                  uuid.UUID
	ClientID, AccessMode, DeploymentMode, TenantPurpose string
	PlatformSubjectID                                   uuid.UUID
	TokenVersion, AuthorizationVersion, PlatformClock   int64
	SecurityEpoch                                       [32]byte
	EnvironmentID, OwnerSubjectID                       uuid.UUID
	OwnerDeploymentMode                                 string
	OwnerPlatformClock                                  int64
}

func (evidence Evidence) validShape() bool {
	return evidence.UserID != uuid.Nil && evidence.SessionID != uuid.Nil &&
		evidence.ClientID != "" && evidence.AccessMode != "" &&
		evidence.TokenVersion > 0 && evidence.AuthorizationVersion > 0 &&
		evidence.PlatformClock > 0 && evidence.SecurityEpoch != [32]byte{} &&
		evidence.PlatformSubjectID != uuid.Nil && evidence.EnvironmentID != uuid.Nil &&
		evidence.OwnerSubjectID != uuid.Nil && evidence.OwnerPlatformClock > 0 &&
		evidence.OwnerDeploymentMode != ""
}
