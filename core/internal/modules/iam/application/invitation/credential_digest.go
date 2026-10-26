package invitation

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"

	"github.com/google/uuid"
)

const (
	credentialPurpose = "invitation_accept_v1"
	hmacDomain        = "harpy:iam-invitation-secret:v1"
	bindingDomain     = "harpy:iam-invitation-route:v1"
)

func invitationSecretHMAC(
	key [32]byte,
	tenantID uuid.UUID,
	selector [16]byte,
	secret [32]byte,
) ([32]byte, error) {
	if key == [32]byte{} || tenantID == uuid.Nil || selector == [16]byte{} ||
		secret == [32]byte{} {
		return [32]byte{}, ErrCredentialInvalid
	}
	hash := hmac.New(sha256.New, key[:])
	writeCredentialPart(hash, []byte(hmacDomain))
	writeCredentialPart(hash, []byte(credentialPurpose))
	writeCredentialPart(hash, tenantID[:])
	writeCredentialPart(hash, selector[:])
	writeCredentialPart(hash, secret[:])
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}

func InvitationBindingSHA256(
	tenantID uuid.UUID,
	selector [16]byte,
) ([32]byte, error) {
	if tenantID == uuid.Nil || selector == [16]byte{} {
		return [32]byte{}, ErrCredentialInvalid
	}
	hash := sha256.New()
	writeCredentialPart(hash, []byte(bindingDomain))
	writeCredentialPart(hash, []byte(credentialPurpose))
	writeCredentialPart(hash, tenantID[:])
	writeCredentialPart(hash, selector[:])
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}

func writeCredentialPart(
	writer interface{ Write([]byte) (int, error) },
	value []byte,
) {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value)))
	_, _ = writer.Write(length[:])
	_, _ = writer.Write(value)
}
