package cryptography

import (
	"context"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type WebAuthnRegistrationExpectation struct {
	Challenge, RPID, Origin  string
	UserVerificationRequired bool
}

type WebAuthnRegistrationResult struct {
	CredentialID, PublicKeyCOSE               []byte
	AAGUID                                    uuid.UUID
	SignCount                                 int64
	BackupEligible, BackupState, Discoverable bool
}

func VerifyWebAuthnRegistration(
	ctx context.Context,
	wire []byte,
	expectation WebAuthnRegistrationExpectation,
) (WebAuthnRegistrationResult, error) {
	if ctx == nil || ctx.Err() != nil || len(wire) < 2 || len(wire) > 64*1024 ||
		len(expectation.Challenge) != 43 || expectation.RPID == "" || expectation.Origin == "" {
		return WebAuthnRegistrationResult{}, ErrRegistrationInvalid
	}
	parsed, err := protocol.ParseCredentialCreationResponseBytes(wire)
	if err != nil || parsed == nil {
		return WebAuthnRegistrationResult{}, ErrRegistrationDenied
	}
	_, err = parsed.Verify(expectation.Challenge, expectation.RPID,
		[]string{expectation.Origin}, []string{expectation.Origin},
		protocol.TopOriginExplicitVerificationMode, false,
		expectation.UserVerificationRequired, true, nil,
		webauthn.CredentialParametersDefault())
	if err != nil {
		return WebAuthnRegistrationResult{}, ErrRegistrationDenied
	}
	auth := parsed.Response.AttestationObject.AuthData
	credentialID := auth.AttData.CredentialID
	publicKey := auth.AttData.CredentialPublicKey
	aaguid, aaguidErr := uuid.FromBytes(auth.AttData.AAGUID)
	discoverable, discoverableErr := webAuthnDiscoverable(parsed.ClientExtensionResults)
	if aaguidErr != nil || discoverableErr != nil || len(credentialID) == 0 ||
		len(credentialID) > 1024 || len(publicKey) == 0 || len(publicKey) > 4096 ||
		auth.Counter > uint32(^uint32(0)>>1) ||
		auth.Flags.HasBackupState() && !auth.Flags.HasBackupEligible() {
		return WebAuthnRegistrationResult{}, ErrRegistrationDenied
	}
	return WebAuthnRegistrationResult{
		CredentialID:  append([]byte(nil), credentialID...),
		PublicKeyCOSE: append([]byte(nil), publicKey...), AAGUID: aaguid,
		SignCount: int64(auth.Counter), BackupEligible: auth.Flags.HasBackupEligible(),
		BackupState:  auth.Flags.HasBackupState(),
		Discoverable: discoverable,
	}, nil
}
