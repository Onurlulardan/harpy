package authentication

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/google/uuid"
)

const refreshLocatorSourceDomain = "harpy:iam-refresh-locator-source:v1"

func RefreshLocatorSourceSHA256(
	tokenID uuid.UUID,
	familyID uuid.UUID,
	selector [16]byte,
	revision int64,
	state string,
	secretHMAC [32]byte,
) ([32]byte, error) {
	if tokenID == uuid.Nil || familyID == uuid.Nil || selector == [16]byte{} ||
		revision <= 0 || !oneOf(state, "active", "consumed", "revoked", "expired") {
		return [32]byte{}, ErrAuthenticationContract
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte(refreshLocatorSourceDomain))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(tokenID[:])
	_, _ = hash.Write(familyID[:])
	_, _ = hash.Write(selector[:])
	var revisionBytes [8]byte
	binary.BigEndian.PutUint64(revisionBytes[:], uint64(revision))
	_, _ = hash.Write(revisionBytes[:])
	_, _ = hash.Write([]byte(state))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(secretHMAC[:])
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}
