package accesstoken

import "github.com/google/uuid"

func (claims Claims) validShape() bool {
	if claims.issuer == "" || claims.audience == "" ||
		!validUUID(claims.subject) || !validUUID(claims.session) ||
		!validUUID(claims.tenant) || !validV4UUID(claims.jti) ||
		claims.issuedAt <= 0 || claims.expiresAt != claims.issuedAt+int64(Lifetime.Seconds()) ||
		claims.authenticationTime <= 0 || claims.authenticationTime > claims.issuedAt ||
		claims.tokenVersion <= 0 || claims.authorizationVersion <= 0 ||
		claims.platformAuthorizationClock < 0 || claims.securityEpoch == [32]byte{} ||
		!validClientMode(claims.client, claims.accessMode) || !validMethods(claims.methods) {
		return false
	}
	if claims.accessMode == AccessModePortal {
		if !claims.hasPortalBinding || !validUUID(claims.portalBinding) {
			return false
		}
	} else if claims.hasPortalBinding {
		return false
	}
	if claims.assurance == AssuranceStepUp {
		return claims.hasStepUpPurpose
	}
	return claims.assurance == AssuranceBaseline && !claims.hasStepUpPurpose
}

func validClientMode(client ClientID, mode AccessMode) bool {
	switch client {
	case ClientPortal:
		return mode == AccessModePortal
	case ClientOperatorConsole:
		return mode == AccessModeOperator
	default:
		return false
	}
}

func validMethods(methods []AuthMethod) bool {
	if len(methods) == 0 || len(methods) > 3 {
		return false
	}
	previous := 0
	for _, method := range methods {
		rank := methodRank(method)
		if rank == 0 || rank <= previous {
			return false
		}
		previous = rank
	}
	return true
}

func methodRank(method AuthMethod) int {
	switch method {
	case AuthMethodPassword:
		return 1
	case AuthMethodWebAuthn:
		return 2
	case AuthMethodTOTP:
		return 3
	default:
		return 0
	}
}

func validUUID(value uuid.UUID) bool {
	return value != uuid.Nil
}

func validV4UUID(value uuid.UUID) bool {
	return validUUID(value) && value.Version() == 4 && value.Variant() == uuid.RFC4122
}
