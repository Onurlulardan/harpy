package authentication

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/google/uuid"
)

const reauthenticationBindingDomain = "harpy:iam-reauth-session:v1"

func ReauthenticationBindingSHA256(
	tenantID, userID, sessionID uuid.UUID,
	tokenVersion, authorizationVersion int64,
	mode AccessMode,
	portalBindingID uuid.UUID,
) ([32]byte, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil || sessionID == uuid.Nil ||
		tokenVersion <= 0 || authorizationVersion <= 0 ||
		(mode != AccessModeOperator && mode != AccessModePortal) ||
		(mode == AccessModeOperator && portalBindingID != uuid.Nil) ||
		(mode == AccessModePortal && portalBindingID == uuid.Nil) {
		return [32]byte{}, ErrAuthenticationContract
	}
	hash := sha256.New()
	writeDigestPart(hash, []byte(reauthenticationBindingDomain))
	writeDigestPart(hash, tenantID[:])
	writeDigestPart(hash, userID[:])
	writeDigestPart(hash, sessionID[:])
	var number [8]byte
	binary.BigEndian.PutUint64(number[:], uint64(tokenVersion))
	writeDigestPart(hash, number[:])
	binary.BigEndian.PutUint64(number[:], uint64(authorizationVersion))
	writeDigestPart(hash, number[:])
	writeDigestPart(hash, []byte(mode))
	writeDigestPart(hash, portalBindingID[:])
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}

func writeDigestPart(hash interface{ Write([]byte) (int, error) }, value []byte) {
	_, _ = hash.Write(value)
	_, _ = hash.Write([]byte{0})
}
