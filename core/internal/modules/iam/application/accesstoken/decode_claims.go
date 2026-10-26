package accesstoken

import "encoding/json"

var requiredClaimNames = []string{
	"iss", "aud", "sub", "sid", "tenant_id", "client_id", "iat", "nbf",
	"exp", "jti", "token_version", "authorization_version", "access_mode",
	"platform_authorization_clock", "security_epoch", "auth_time", "acr", "amr",
}

func decodeClaims(document []byte) (Claims, bool) {
	fields, ok := decodeStrictObject(document)
	if !ok || !validClaimNames(fields) {
		return Claims{}, false
	}
	claims := Claims{}
	if !decodeClaimScalars(fields, &claims) || !decodeClaimIdentity(fields, &claims) ||
		!decodeClaimOptionals(fields, &claims) || !claims.validShape() {
		return Claims{}, false
	}
	return claims, true
}

func validClaimNames(fields map[string]json.RawMessage) bool {
	if len(fields) < len(requiredClaimNames) || len(fields) > len(requiredClaimNames)+2 {
		return false
	}
	allowed := make(map[string]struct{}, len(requiredClaimNames)+2)
	for _, name := range requiredClaimNames {
		allowed[name] = struct{}{}
		if _, exists := fields[name]; !exists {
			return false
		}
	}
	allowed["portal_binding_id"] = struct{}{}
	allowed["step_up_purpose_hash"] = struct{}{}
	for name := range fields {
		if _, exists := allowed[name]; !exists {
			return false
		}
	}
	return true
}

func decodeClaimScalars(fields map[string]json.RawMessage, claims *Claims) bool {
	var ok bool
	claims.issuer, ok = decodeString(fields["iss"])
	if !ok {
		return false
	}
	claims.audience, ok = decodeString(fields["aud"])
	if !ok {
		return false
	}
	client, ok := decodeString(fields["client_id"])
	claims.client = ClientID(client)
	mode, modeOK := decodeString(fields["access_mode"])
	claims.accessMode = AccessMode(mode)
	assurance, assuranceOK := decodeString(fields["acr"])
	claims.assurance = Assurance(assurance)
	claims.methods, ok = decodeMethods(fields["amr"])
	return ok && modeOK && assuranceOK
}

func decodeClaimIdentity(fields map[string]json.RawMessage, claims *Claims) bool {
	var ok bool
	claims.subject, ok = decodeCanonicalUUID(fields["sub"], false)
	if !ok {
		return false
	}
	claims.session, ok = decodeCanonicalUUID(fields["sid"], false)
	if !ok {
		return false
	}
	claims.tenant, ok = decodeCanonicalUUID(fields["tenant_id"], false)
	if !ok {
		return false
	}
	claims.jti, ok = decodeCanonicalUUID(fields["jti"], true)
	return ok && decodeClaimTimesAndVersions(fields, claims)
}

func decodeClaimTimesAndVersions(fields map[string]json.RawMessage, claims *Claims) bool {
	var ok bool
	claims.issuedAt, ok = decodePositiveInt64(fields["iat"])
	if !ok {
		return false
	}
	notBefore, ok := decodePositiveInt64(fields["nbf"])
	if !ok || notBefore != claims.issuedAt {
		return false
	}
	claims.expiresAt, ok = decodePositiveInt64(fields["exp"])
	if !ok {
		return false
	}
	claims.tokenVersion, ok = decodePositiveInt64(fields["token_version"])
	if !ok {
		return false
	}
	claims.authorizationVersion, ok = decodePositiveInt64(fields["authorization_version"])
	if !ok {
		return false
	}
	claims.platformAuthorizationClock, ok = decodeNonNegativeInt64(
		fields["platform_authorization_clock"],
	)
	if !ok {
		return false
	}
	epoch, ok := decodeString(fields["security_epoch"])
	if !ok {
		return false
	}
	decodedEpoch, ok := decodeBase64URL(epoch, len(claims.securityEpoch))
	if !ok {
		return false
	}
	copy(claims.securityEpoch[:], decodedEpoch)
	claims.authenticationTime, ok = decodePositiveInt64(fields["auth_time"])
	return ok
}

func decodeClaimOptionals(fields map[string]json.RawMessage, claims *Claims) bool {
	if raw, exists := fields["portal_binding_id"]; exists {
		value, ok := decodeCanonicalUUID(raw, false)
		if !ok {
			return false
		}
		claims.portalBinding, claims.hasPortalBinding = value, true
	}
	if raw, exists := fields["step_up_purpose_hash"]; exists {
		value, ok := decodeString(raw)
		if !ok {
			return false
		}
		decoded, ok := decodeBase64URL(value, len(claims.stepUpPurposeHash))
		if !ok {
			return false
		}
		copy(claims.stepUpPurposeHash[:], decoded)
		claims.hasStepUpPurpose = true
	}
	return true
}
