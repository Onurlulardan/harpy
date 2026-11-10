package authentication

import (
	"time"

	"github.com/google/uuid"
)

type ClientID string

const ClientOperatorConsole ClientID = "harpy-operator-console-v1"

type AccessMode string

const AccessModeOperator AccessMode = "operator_console"

type Assurance string

const (
	AssuranceBaseline Assurance = "urn:harpy:acr:baseline"
	AssuranceStepUp   Assurance = "urn:harpy:acr:step-up"
)

type AuthMethod string

const (
	AuthMethodPassword AuthMethod = "pwd"
	AuthMethodWebAuthn AuthMethod = "webauthn"
	AuthMethodTOTP     AuthMethod = "totp"
)

type AccessClaims struct {
	Subject, Session, Tenant, JTI      uuid.UUID
	Client                             ClientID
	IssuedAt, AuthenticationTime       time.Time
	TokenVersion, AuthorizationVersion int64
	PlatformAuthorizationClock         int64
	SecurityEpoch                      [32]byte
	AccessMode                         AccessMode
	PortalBinding                      *uuid.UUID
	Assurance                          Assurance
	Methods                            []AuthMethod
	StepUpPurposeHash                  []byte
}

func clientMatchesMode(client ClientID, mode AccessMode) bool {
	return mode == AccessModeOperator && client == ClientOperatorConsole
}
