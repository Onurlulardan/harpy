package accesstoken

import "time"

const (
	Algorithm       = "Ed25519"
	Type            = "application/vnd.harpy.access+jwt"
	MaximumWireSize = 8192
	Lifetime        = 15 * time.Minute
	ClockSkew       = 60 * time.Second
)

type ClientID string

const (
	ClientOperatorConsole ClientID = "harpy-operator-console-v1"
	ClientPortal          ClientID = "harpy-portal-v1"
)

type AccessMode string

const (
	AccessModeOperator AccessMode = "operator_console"
	AccessModePortal   AccessMode = "portal"
)

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
