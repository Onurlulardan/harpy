package authentication

import (
	"crypto/subtle"

	"github.com/google/uuid"
)

type RecoveryRequest struct{ Code RecoveryCode }

func (request RecoveryRequest) validate() error {
	if nilInterface(request.Code) {
		return ErrAuthenticationContract
	}
	return nil
}

type RecoveryLookup struct {
	HMACKeyID string
	CodeHMAC  [32]byte
}

type RecoveryCredential struct {
	CodeID, AuthenticatorID, UserID                         uuid.UUID
	CodeRevision, AuthenticatorRevision, UserRevision       int64
	TokenVersion                                            int64
	CodeState, AuthenticatorState, UserStatus, IdentityKind string
	HMACKeyID                                               string
	CodeHMAC                                                [32]byte
}

func (credential RecoveryCredential) validate(lookup RecoveryLookup) error {
	if credential.CodeID == uuid.Nil || credential.AuthenticatorID == uuid.Nil ||
		credential.UserID == uuid.Nil || credential.CodeRevision <= 0 ||
		credential.AuthenticatorRevision <= 0 || credential.UserRevision <= 0 ||
		credential.TokenVersion <= 0 || credential.CodeState != "active" ||
		credential.AuthenticatorState != "active" || credential.UserStatus != "active" ||
		credential.IdentityKind != "human" || credential.HMACKeyID != lookup.HMACKeyID ||
		subtle.ConstantTimeCompare(credential.CodeHMAC[:], lookup.CodeHMAC[:]) != 1 {
		return ErrAuthenticationStorage
	}
	return nil
}

type RecoveryConsumption struct {
	CodeID, AuthenticatorID, UserID                     uuid.UUID
	ExpectedCodeRevision, ExpectedAuthenticatorRevision int64
	ExpectedUserRevision, ExpectedTokenVersion          int64
	HMACKeyID                                           string
	CodeHMAC                                            [32]byte
}

type RecoveryConsumed struct{ TokenVersion int64 }

func (consumed RecoveryConsumed) validate(previous int64) error {
	if consumed.TokenVersion != previous+1 {
		return ErrAuthenticationStorage
	}
	return nil
}

func (credential RecoveryCredential) consumption() RecoveryConsumption {
	return RecoveryConsumption{
		CodeID: credential.CodeID, AuthenticatorID: credential.AuthenticatorID,
		UserID: credential.UserID, ExpectedCodeRevision: credential.CodeRevision,
		ExpectedAuthenticatorRevision: credential.AuthenticatorRevision,
		ExpectedUserRevision:          credential.UserRevision,
		ExpectedTokenVersion:          credential.TokenVersion,
		HMACKeyID:                     credential.HMACKeyID, CodeHMAC: credential.CodeHMAC,
	}
}
