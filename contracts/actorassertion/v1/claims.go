package actorassertionv1

import (
	"regexp"

	"github.com/google/uuid"
)

var keyIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,62}$`)

type Claims struct {
	KeyID                          string
	Decision                       DecisionEvidence
	DecisionSHA256                 [32]byte
	JTI                            uuid.UUID
	IssuedAt, NotBefore, ExpiresAt int64
}

func newClaims(keyID string, decision DecisionEvidence, jti uuid.UUID, issuedAt int64) (Claims, error) {
	digest, err := decision.Digest()
	if err != nil {
		return Claims{}, err
	}
	claims := Claims{KeyID: keyID, Decision: decision, DecisionSHA256: digest,
		JTI: jti, IssuedAt: issuedAt, NotBefore: issuedAt,
		ExpiresAt: issuedAt + LifetimeSeconds}
	if claims.Validate() != nil {
		return Claims{}, ErrContract
	}
	return claims, nil
}

func (claims Claims) Validate() error {
	digest, err := claims.Decision.Digest()
	if err != nil || !keyIDPattern.MatchString(claims.KeyID) || claims.JTI == uuid.Nil ||
		claims.DecisionSHA256 != digest || claims.IssuedAt <= 0 ||
		claims.NotBefore != claims.IssuedAt ||
		claims.ExpiresAt-claims.IssuedAt != LifetimeSeconds {
		return ErrContract
	}
	return nil
}
