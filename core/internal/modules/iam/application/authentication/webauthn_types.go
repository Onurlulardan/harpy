package authentication

import (
	"crypto/sha256"
	"crypto/subtle"
)

type WebAuthnExpectation struct {
	Challenge                []byte
	CredentialID             []byte
	PublicKeyCOSE            []byte
	RPID                     string
	Origin                   string
	UserHandle               []byte
	UserVerificationRequired bool
}

type WebAuthnResult struct {
	SignCount      uint32
	BackupEligible bool
	BackupState    bool
	UserVerified   bool
}

type WebAuthnCredential struct {
	CredentialID             []byte
	PublicKeyCOSE            []byte
	RPID                     string
	RPIDSHA256               [32]byte
	OriginPolicyVersion      int64
	SignCount                int64
	BackupEligible           bool
	BackupState              bool
	UserVerificationRequired bool
	Revision                 int64
}

func (credential WebAuthnCredential) valid() bool {
	rpIDHash := sha256.Sum256([]byte(credential.RPID))
	return len(credential.CredentialID) >= 1 && len(credential.CredentialID) <= 1024 &&
		len(credential.PublicKeyCOSE) >= 1 && len(credential.PublicKeyCOSE) <= 4096 &&
		credential.RPID != "" && credential.RPIDSHA256 != [32]byte{} &&
		subtle.ConstantTimeCompare(rpIDHash[:], credential.RPIDSHA256[:]) == 1 &&
		credential.OriginPolicyVersion > 0 && credential.SignCount >= 0 &&
		credential.Revision > 0 && (!credential.BackupState || credential.BackupEligible)
}
