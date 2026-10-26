package accesstoken

import (
	"context"
	"crypto/ed25519"
	"time"
)

type Verifier struct {
	profile  Profile
	resolver KeyResolver
}

func NewVerifier(profile Profile, resolver KeyResolver) (*Verifier, error) {
	if profile.issuer == "" || profile.audience == "" || resolver == nil {
		return nil, ErrInvalidProfile
	}
	return &Verifier{profile: profile, resolver: resolver}, nil
}

func (verifier *Verifier) Verify(
	ctx context.Context,
	token string,
	now time.Time,
) (Claims, error) {
	if verifier == nil || verifier.resolver == nil || now.IsZero() {
		return Claims{}, ErrInvalidToken
	}
	segments, ok := splitCompact(token)
	if !ok {
		return Claims{}, ErrInvalidToken
	}
	header, ok := decodeSegment(segments.header)
	if !ok {
		return Claims{}, ErrInvalidToken
	}
	keyID, ok := decodeHeader(header)
	if !ok {
		return Claims{}, ErrInvalidToken
	}
	publicKey, err := verifier.resolver.AccessTokenKey(ctx, keyID)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return Claims{}, ErrInvalidToken
	}
	signature, ok := decodeBase64URL(segments.signature, ed25519.SignatureSize)
	if !ok || !ed25519.Verify(publicKey, []byte(segments.header+"."+segments.payload), signature) {
		return Claims{}, ErrInvalidToken
	}
	payload, ok := decodeSegment(segments.payload)
	if !ok {
		return Claims{}, ErrInvalidToken
	}
	claims, ok := decodeClaims(payload)
	if !ok || !verifier.validClaims(claims, now.Unix()) {
		return Claims{}, ErrInvalidToken
	}
	return claims, nil
}

func (verifier *Verifier) validClaims(claims Claims, now int64) bool {
	skew := int64(ClockSkew.Seconds())
	return claims.issuer == verifier.profile.issuer &&
		claims.audience == verifier.profile.audience &&
		now > 0 && claims.issuedAt <= now+skew &&
		now < claims.expiresAt+skew
}

func decodeSegment(segment string) ([]byte, bool) {
	decoded, err := base64URLStrict(segment)
	return decoded, err == nil
}
