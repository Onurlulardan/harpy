package authentication

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"

	"github.com/google/uuid"
)

const (
	initialPasswordPurpose = "installation_initial_password_v1"
	initialSecretDomain    = "harpy:iam-initial-password-continuation:v1"
	initialCSRFDomain      = "harpy:iam-initial-password-csrf:v1"
	initialRouteDomain     = "harpy:iam-initial-password-route:v1"
	initialCompleteDomain  = "harpy:iam-initial-password-complete:v1"
)

func initialCompletionRequestSHA256(continuation InitialPasswordContinuation) [32]byte {
	hash := sha256.New()
	for _, part := range [][]byte{[]byte(initialCompleteDomain),
		continuation.TenantID[:], continuation.UserID[:],
		continuation.InstallationOperationID[:], continuation.CredentialID[:],
		int64Bytes(continuation.CredentialVersion), continuation.Selector[:]} {
		writeInitialPart(hash, part)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func initialSecretHMAC(
	key [32]byte,
	tenantID, userID, operationID, credentialID uuid.UUID,
	credentialVersion int64,
	selector [16]byte,
	secret [32]byte,
) [32]byte {
	mac := hmac.New(sha256.New, key[:])
	for _, part := range [][]byte{[]byte(initialSecretDomain), []byte(initialPasswordPurpose),
		tenantID[:], userID[:], operationID[:], credentialID[:],
		int64Bytes(credentialVersion), selector[:], secret[:]} {
		writeInitialPart(mac, part)
	}
	var result [32]byte
	copy(result[:], mac.Sum(nil))
	return result
}

func initialCSRFHMAC(key [32]byte, selector [16]byte, csrf [32]byte) [32]byte {
	mac := hmac.New(sha256.New, key[:])
	for _, part := range [][]byte{
		[]byte(initialCSRFDomain), []byte(initialPasswordPurpose), selector[:], csrf[:],
	} {
		writeInitialPart(mac, part)
	}
	var result [32]byte
	copy(result[:], mac.Sum(nil))
	return result
}

func InitialPasswordRouteBinding(tenantID uuid.UUID, selector [16]byte) [32]byte {
	hash := sha256.New()
	for _, part := range [][]byte{
		[]byte(initialRouteDomain), []byte(initialPasswordPurpose), tenantID[:], selector[:],
	} {
		writeInitialPart(hash, part)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func int64Bytes(value int64) []byte {
	var result [8]byte
	binary.BigEndian.PutUint64(result[:], uint64(value))
	return result[:]
}

func writeInitialPart(writer interface{ Write([]byte) (int, error) }, value []byte) {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value)))
	_, _ = writer.Write(length[:])
	_, _ = writer.Write(value)
}
