package accesstoken

func NewClaims(profile Profile, input ClaimsInput) (Claims, error) {
	if profile.issuer == "" || profile.audience == "" ||
		input.IssuedAt.Nanosecond() != 0 || input.AuthenticationTime.Nanosecond() != 0 {
		return Claims{}, ErrInvalidClaims
	}
	claims := Claims{
		issuer: profile.issuer, audience: profile.audience,
		subject: input.Subject, session: input.Session, tenant: input.Tenant,
		client: input.Client, issuedAt: input.IssuedAt.Unix(),
		expiresAt: input.IssuedAt.Add(Lifetime).Unix(), jti: input.JTI,
		tokenVersion:               input.TokenVersion,
		authorizationVersion:       input.AuthorizationVersion,
		platformAuthorizationClock: input.PlatformAuthorizationClock,
		securityEpoch:              input.SecurityEpoch,
		accessMode:                 input.AccessMode, authenticationTime: input.AuthenticationTime.Unix(),
		assurance: input.Assurance, methods: append([]AuthMethod(nil), input.Methods...),
	}
	if input.PortalBinding != nil {
		claims.portalBinding = *input.PortalBinding
		claims.hasPortalBinding = true
	}
	if len(input.StepUpPurposeHash) == len(claims.stepUpPurposeHash) {
		copy(claims.stepUpPurposeHash[:], input.StepUpPurposeHash)
		claims.hasStepUpPurpose = true
	} else if len(input.StepUpPurposeHash) != 0 {
		return Claims{}, ErrInvalidClaims
	}
	if !claims.validShape() {
		return Claims{}, ErrInvalidClaims
	}
	return claims, nil
}
