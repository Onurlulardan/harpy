package authentication

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/google/uuid"
)

const reauthenticationContextDomain = "harpy:iam-reauth-context:v1"

func ReauthenticationContextSHA256(
	tenantID, userID, sessionID uuid.UUID,
	purpose [32]byte,
	actionPolicyVersion int64,
	method ReauthenticationMethod,
	originSHA256 [32]byte,
) ([32]byte, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil || sessionID == uuid.Nil ||
		purpose == [32]byte{} || actionPolicyVersion <= 0 || !method.valid() ||
		originSHA256 == [32]byte{} {
		return [32]byte{}, ErrAuthenticationContract
	}
	hash := sha256.New()
	writeDigestPart(hash, []byte(reauthenticationContextDomain))
	writeDigestPart(hash, tenantID[:])
	writeDigestPart(hash, userID[:])
	writeDigestPart(hash, sessionID[:])
	writeDigestPart(hash, purpose[:])
	var number [8]byte
	binary.BigEndian.PutUint64(number[:], uint64(actionPolicyVersion))
	writeDigestPart(hash, number[:])
	writeDigestPart(hash, []byte(method))
	writeDigestPart(hash, originSHA256[:])
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}
