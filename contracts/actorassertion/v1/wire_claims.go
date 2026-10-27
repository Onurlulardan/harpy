package actorassertionv1

import (
	"encoding/base64"

	"github.com/google/uuid"
)

type wireClaims struct {
	Contract                 string `json:"contract"`
	Algorithm                string `json:"algorithm"`
	KeyID                    string `json:"key_id"`
	Issuer                   string `json:"iss"`
	Audience                 string `json:"aud"`
	TenantID                 string `json:"tenant_id"`
	SubjectID                string `json:"sub"`
	PermissionID             string `json:"permission_id"`
	PermissionResource       string `json:"permission_resource"`
	PermissionAction         string `json:"permission_action"`
	TargetDomain             string `json:"target_domain"`
	TargetBindingSHA256      string `json:"target_binding_sha256"`
	PermissionDecisionSHA256 string `json:"permission_decision_sha256"`
	IAMGroupGeneration       int64  `json:"iam_group_generation"`
	IAMGeneration            int64  `json:"iam_generation"`
	TargetGeneration         int64  `json:"target_generation"`
	TokenVersion             int64  `json:"token_version"`
	AuthorizationVersion     int64  `json:"authorization_version"`
	AuthorizationClock       int64  `json:"authorization_clock"`
	JTI                      string `json:"jti"`
	IssuedAt                 int64  `json:"iat"`
	NotBefore                int64  `json:"nbf"`
	ExpiresAt                int64  `json:"exp"`
}

func claimsToWire(claims Claims) wireClaims {
	decision := claims.Decision
	return wireClaims{Contract: ContractVersion, Algorithm: Algorithm,
		KeyID: claims.KeyID, Issuer: Issuer, Audience: Audience(decision.TargetDomain),
		TenantID: decision.TenantID.String(), SubjectID: decision.SubjectID.String(),
		PermissionID: decision.PermissionID.String(), PermissionResource: decision.ResourceKey,
		PermissionAction: decision.ActionKey, TargetDomain: string(decision.TargetDomain),
		TargetBindingSHA256:      encodeDigest(decision.TargetBindingSHA256),
		PermissionDecisionSHA256: encodeDigest(claims.DecisionSHA256),
		IAMGroupGeneration:       decision.IAMGroupGeneration, IAMGeneration: decision.IAMGeneration,
		TargetGeneration: decision.TargetGeneration, TokenVersion: decision.TokenVersion,
		AuthorizationVersion: decision.AuthorizationVersion,
		AuthorizationClock:   decision.AuthorizationClock, JTI: claims.JTI.String(),
		IssuedAt: claims.IssuedAt, NotBefore: claims.NotBefore, ExpiresAt: claims.ExpiresAt}
}

func (wire wireClaims) typed() (Claims, error) {
	tenantID, tenantErr := uuid.Parse(wire.TenantID)
	subjectID, subjectErr := uuid.Parse(wire.SubjectID)
	permissionID, permissionErr := uuid.Parse(wire.PermissionID)
	jti, jtiErr := uuid.Parse(wire.JTI)
	targetDigest, targetErr := decodeDigest(wire.TargetBindingSHA256)
	decisionDigest, decisionErr := decodeDigest(wire.PermissionDecisionSHA256)
	if tenantErr != nil || subjectErr != nil || permissionErr != nil || jtiErr != nil ||
		targetErr != nil || decisionErr != nil || wire.Contract != ContractVersion ||
		wire.Algorithm != Algorithm || wire.Issuer != Issuer {
		return Claims{}, ErrContract
	}
	decision := DecisionEvidence{TenantID: tenantID, SubjectID: subjectID,
		PermissionID: permissionID, TargetDomain: Domain(wire.TargetDomain),
		ResourceKey: wire.PermissionResource, ActionKey: wire.PermissionAction,
		TargetBindingSHA256: targetDigest, IAMGroupGeneration: wire.IAMGroupGeneration,
		IAMGeneration: wire.IAMGeneration, TargetGeneration: wire.TargetGeneration,
		TokenVersion: wire.TokenVersion, AuthorizationVersion: wire.AuthorizationVersion,
		AuthorizationClock: wire.AuthorizationClock}
	claims := Claims{KeyID: wire.KeyID, Decision: decision, DecisionSHA256: decisionDigest,
		JTI: jti, IssuedAt: wire.IssuedAt, NotBefore: wire.NotBefore, ExpiresAt: wire.ExpiresAt}
	if wire.Audience != Audience(decision.TargetDomain) || claims.Validate() != nil {
		return Claims{}, ErrContract
	}
	return claims, nil
}

func encodeDigest(value [32]byte) string {
	return base64.RawURLEncoding.EncodeToString(value[:])
}

func decodeDigest(value string) ([32]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) != 32 || base64.RawURLEncoding.EncodeToString(decoded) != value {
		return [32]byte{}, ErrContract
	}
	var result [32]byte
	copy(result[:], decoded)
	return result, nil
}
