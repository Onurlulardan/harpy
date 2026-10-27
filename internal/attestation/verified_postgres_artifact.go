package attestation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type postgresArtifactContract interface {
	Validate() error
}

type VerifiedPostgresArtifact[T postgresArtifactContract] struct {
	value        T
	canonical    []byte
	digestSHA256 string
	maximumBytes int
	verified     bool
}

func (artifact VerifiedPostgresArtifact[T]) Value() T {
	if !artifact.verified {
		var zero T
		return zero
	}
	var cloned T
	if err := decodeStrictBounded(artifact.canonical, &cloned, artifact.maximumBytes); err != nil {
		var zero T
		return zero
	}
	return cloned
}

func (artifact VerifiedPostgresArtifact[T]) CanonicalJSON() []byte {
	return append([]byte(nil), artifact.canonical...)
}

func (artifact VerifiedPostgresArtifact[T]) DigestSHA256() string {
	return artifact.digestSHA256
}

func (artifact VerifiedPostgresArtifact[T]) IsVerified() bool {
	return artifact.verified
}

func verifyPostgresArtifact[T postgresArtifactContract](
	document []byte,
	maximumBytes int,
) (VerifiedPostgresArtifact[T], error) {
	var value T
	if err := decodeStrictBounded(document, &value, maximumBytes); err != nil {
		return VerifiedPostgresArtifact[T]{}, fmt.Errorf("decode PostgreSQL artifact: %w", err)
	}
	if err := value.Validate(); err != nil {
		return VerifiedPostgresArtifact[T]{}, fmt.Errorf("validate PostgreSQL artifact: %w", err)
	}
	canonical, err := canonicalizeBounded(document, maximumBytes)
	if err != nil {
		return VerifiedPostgresArtifact[T]{}, err
	}
	digest := sha256.Sum256(canonical)
	return VerifiedPostgresArtifact[T]{
		value: value, canonical: canonical,
		digestSHA256: hex.EncodeToString(digest[:]),
		maximumBytes: maximumBytes, verified: true,
	}, nil
}
