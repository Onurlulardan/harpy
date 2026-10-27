package gcpkms

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/google/uuid"
)

const totpAADDomain = "harpy:iam:totp:aes-256-gcm:v1"

func TOTPAdditionalAuthenticatedData(
	environmentID, tenantID, userID, authenticatorID uuid.UUID,
	keyID string,
) ([32]byte, error) {
	if environmentID == uuid.Nil || tenantID == uuid.Nil || userID == uuid.Nil ||
		authenticatorID == uuid.Nil {
		return [32]byte{}, ErrContract
	}
	if _, err := parseKeyVersion(keyID); err != nil {
		return [32]byte{}, ErrContract
	}
	hash := sha256.New()
	for _, part := range [][]byte{
		[]byte(totpAADDomain), environmentID[:], tenantID[:], userID[:],
		authenticatorID[:], []byte("rfc6238_sha1_6_30_v1"),
		[]byte("aes_256_gcm_v1"), []byte(keyID),
	} {
		var length [4]byte
		binary.BigEndian.PutUint32(length[:], uint32(len(part)))
		_, _ = hash.Write(length[:])
		_, _ = hash.Write(part)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}
