package accesstoken

import "github.com/google/uuid"

type Claims struct {
	issuer, audience           string
	subject, session           uuid.UUID
	tenant, jti                uuid.UUID
	client                     ClientID
	issuedAt, expiresAt        int64
	tokenVersion               int64
	authorizationVersion       int64
	platformAuthorizationClock int64
	securityEpoch              [32]byte
	accessMode                 AccessMode
	portalBinding              uuid.UUID
	hasPortalBinding           bool
	authenticationTime         int64
	assurance                  Assurance
	methods                    []AuthMethod
	stepUpPurposeHash          [32]byte
	hasStepUpPurpose           bool
}

func (claims Claims) Issuer() string              { return claims.issuer }
func (claims Claims) Audience() string            { return claims.audience }
func (claims Claims) Subject() uuid.UUID          { return claims.subject }
func (claims Claims) Session() uuid.UUID          { return claims.session }
func (claims Claims) Tenant() uuid.UUID           { return claims.tenant }
func (claims Claims) JTI() uuid.UUID              { return claims.jti }
func (claims Claims) Client() ClientID            { return claims.client }
func (claims Claims) IssuedAtUnix() int64         { return claims.issuedAt }
func (claims Claims) ExpiresAtUnix() int64        { return claims.expiresAt }
func (claims Claims) TokenVersion() int64         { return claims.tokenVersion }
func (claims Claims) AuthorizationVersion() int64 { return claims.authorizationVersion }
func (claims Claims) PlatformAuthorizationClock() int64 {
	return claims.platformAuthorizationClock
}
func (claims Claims) SecurityEpoch() [32]byte       { return claims.securityEpoch }
func (claims Claims) AccessMode() AccessMode        { return claims.accessMode }
func (claims Claims) AuthenticationTimeUnix() int64 { return claims.authenticationTime }
func (claims Claims) Assurance() Assurance          { return claims.assurance }
func (claims Claims) Methods() []AuthMethod         { return append([]AuthMethod(nil), claims.methods...) }

func (claims Claims) PortalBinding() (uuid.UUID, bool) {
	return claims.portalBinding, claims.hasPortalBinding
}

func (claims Claims) StepUpPurposeHash() ([32]byte, bool) {
	return claims.stepUpPurposeHash, claims.hasStepUpPurpose
}
