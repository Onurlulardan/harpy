package identity

import (
	"crypto/sha256"
	"regexp"
)

var lookupKeyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,127}$`)

type DirectoryProjection struct {
	display       string
	normalized    string
	normalizedSHA [sha256.Size]byte
	lookupKeyID   string
	hmac          [sha256.Size]byte
}

func NewDirectoryProjection(
	display string, normalized string, normalizedSHA [sha256.Size]byte,
	lookupKeyID string, hmac [sha256.Size]byte,
) (DirectoryProjection, error) {
	if display == "" || len(normalized) < 3 || len(normalized) > 254 ||
		sha256.Sum256([]byte(normalized)) != normalizedSHA ||
		!lookupKeyPattern.MatchString(lookupKeyID) || hmac == [sha256.Size]byte{} {
		return DirectoryProjection{}, ErrInvalidCommand
	}
	return DirectoryProjection{
		display: display, normalized: normalized, normalizedSHA: normalizedSHA,
		lookupKeyID: lookupKeyID, hmac: hmac,
	}, nil
}

func (projection DirectoryProjection) Valid() bool {
	_, err := NewDirectoryProjection(
		projection.display, projection.normalized, projection.normalizedSHA,
		projection.lookupKeyID, projection.hmac,
	)
	return err == nil
}

func (projection DirectoryProjection) Display() string            { return projection.display }
func (projection DirectoryProjection) Normalized() string         { return projection.normalized }
func (projection DirectoryProjection) NormalizedSHA256() [32]byte { return projection.normalizedSHA }
func (projection DirectoryProjection) LookupKeyID() string        { return projection.lookupKeyID }
func (projection DirectoryProjection) HMAC() [sha256.Size]byte    { return projection.hmac }
