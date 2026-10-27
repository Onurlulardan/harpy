package lifecycledeliveryv1

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"

	kmsv1 "github.com/Onurlulardan/harpy/contracts/kms/v1"
	"github.com/google/uuid"
)

const (
	NonceBytes      = 12
	CiphertextBytes = 66
	TagBytes        = 16
	PlaintextBytes  = 66
)

const aadDomain = "harpy:iam-lifecycle-delivery:v1"

var ErrEnvelope = errors.New("IAM lifecycle delivery envelope is invalid")

type Envelope struct {
	ArtifactID, SubjectUserID, ResourceID, TokenID uuid.UUID
	Purpose                                        Purpose
	KeyID                                          string
	Nonce                                          [NonceBytes]byte
	Ciphertext                                     [CiphertextBytes]byte
	Tag                                            [TagBytes]byte
	AADSHA256, CiphertextSHA256                    [32]byte
}

func (envelope Envelope) Validate(tenantID uuid.UUID) error {
	if tenantID == uuid.Nil || envelope.ArtifactID == uuid.Nil ||
		envelope.SubjectUserID == uuid.Nil || envelope.ResourceID == uuid.Nil ||
		envelope.TokenID == uuid.Nil || envelope.Purpose.Validate() != nil ||
		envelope.Nonce == [NonceBytes]byte{} || envelope.Ciphertext == [CiphertextBytes]byte{} ||
		envelope.Tag == [TagBytes]byte{} {
		return ErrEnvelope
	}
	if _, err := kmsv1.ParseGoogleCloudKeyVersion(envelope.KeyID); err != nil {
		return ErrEnvelope
	}
	want, err := AADSHA256(tenantID, envelope.ArtifactID, envelope.SubjectUserID,
		envelope.ResourceID, envelope.TokenID, envelope.Purpose, envelope.KeyID)
	if err != nil || want != envelope.AADSHA256 ||
		sealedSHA256(envelope.Ciphertext, envelope.Tag) != envelope.CiphertextSHA256 {
		return ErrEnvelope
	}
	return nil
}

func AADSHA256(tenantID, artifactID, subjectUserID, resourceID, tokenID uuid.UUID,
	purpose Purpose, keyID string) ([32]byte, error) {
	if tenantID == uuid.Nil || artifactID == uuid.Nil || subjectUserID == uuid.Nil ||
		resourceID == uuid.Nil || tokenID == uuid.Nil || purpose.Validate() != nil {
		return [32]byte{}, ErrEnvelope
	}
	if _, err := kmsv1.ParseGoogleCloudKeyVersion(keyID); err != nil {
		return [32]byte{}, ErrEnvelope
	}
	hash := sha256.New()
	for _, part := range [][]byte{[]byte(aadDomain), tenantID[:], artifactID[:],
		subjectUserID[:], resourceID[:], tokenID[:], []byte(purpose), []byte(keyID)} {
		var length [4]byte
		binary.BigEndian.PutUint32(length[:], uint32(len(part)))
		_, _ = hash.Write(length[:])
		_, _ = hash.Write(part)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}

func NewEnvelope(artifactID, subjectUserID, resourceID, tokenID uuid.UUID,
	purpose Purpose, keyID string, nonce [NonceBytes]byte,
	ciphertext [CiphertextBytes]byte, tag [TagBytes]byte, aad [32]byte) Envelope {
	return Envelope{ArtifactID: artifactID, SubjectUserID: subjectUserID,
		ResourceID: resourceID, TokenID: tokenID, Purpose: purpose, KeyID: keyID,
		Nonce: nonce, Ciphertext: ciphertext, Tag: tag, AADSHA256: aad,
		CiphertextSHA256: sealedSHA256(ciphertext, tag)}
}

func sealedSHA256(ciphertext [CiphertextBytes]byte, tag [TagBytes]byte) [32]byte {
	hash := sha256.New()
	_, _ = hash.Write(ciphertext[:])
	_, _ = hash.Write(tag[:])
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}
