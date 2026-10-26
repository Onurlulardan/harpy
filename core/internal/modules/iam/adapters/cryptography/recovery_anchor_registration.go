package cryptography

import (
	"context"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type RecoveryAnchorRegistrationExpectation struct {
	Challenge      string
	RPID           string
	AllowedOrigins []string
}

type RecoveryAnchorRegistrationResult struct {
	CredentialID, PublicKeyCOSE        []byte
	AAGUID                             uuid.UUID
	AttestationFormat, AttestationType string
	SignCount                          int64
}

func VerifyRecoveryAnchorRegistration(
	ctx context.Context,
	wire []byte,
	expectation RecoveryAnchorRegistrationExpectation,
) (RecoveryAnchorRegistrationResult, error) {
	if ctx == nil || ctx.Err() != nil || len(wire) < 2 || len(wire) > 64*1024 ||
		len(expectation.Challenge) != 43 || expectation.RPID == "" ||
		len(expectation.AllowedOrigins) == 0 || len(expectation.AllowedOrigins) > 32 {
		return RecoveryAnchorRegistrationResult{}, ErrRegistrationInvalid
	}
	parsed, err := protocol.ParseCredentialCreationResponseBytes(wire)
	if err != nil || parsed == nil {
		return RecoveryAnchorRegistrationResult{}, ErrRegistrationDenied
	}
	clientHash, err := parsed.Verify(expectation.Challenge, expectation.RPID,
		expectation.AllowedOrigins, expectation.AllowedOrigins,
		protocol.TopOriginExplicitVerificationMode, false, true, true, nil,
		webauthn.CredentialParametersDefault())
	if err != nil {
		return RecoveryAnchorRegistrationResult{}, ErrRegistrationDenied
	}
	credential, err := webauthn.NewCredential(clientHash, parsed)
	if err != nil || credential == nil {
		return RecoveryAnchorRegistrationResult{}, ErrRegistrationDenied
	}
	discoverable, discoverableErr := webAuthnDiscoverable(parsed.ClientExtensionResults)
	aaguid, aaguidErr := uuid.FromBytes(credential.Authenticator.AAGUID)
	if discoverableErr != nil || !discoverable || aaguidErr != nil || aaguid == uuid.Nil ||
		len(credential.ID) == 0 || len(credential.ID) > 1024 ||
		len(credential.PublicKey) == 0 || len(credential.PublicKey) > 4096 ||
		credential.Authenticator.SignCount > uint32(^uint32(0)>>1) ||
		credential.Flags.BackupEligible || credential.Flags.BackupState ||
		strings.TrimSpace(credential.AttestationFormat) == "" ||
		credential.AttestationFormat == "none" ||
		strings.TrimSpace(credential.AttestationType) == "" ||
		credential.AttestationType == "none" {
		return RecoveryAnchorRegistrationResult{}, ErrRegistrationDenied
	}
	return RecoveryAnchorRegistrationResult{
		CredentialID:  append([]byte(nil), credential.ID...),
		PublicKeyCOSE: append([]byte(nil), credential.PublicKey...),
		AAGUID:        aaguid, AttestationFormat: credential.AttestationFormat,
		AttestationType: credential.AttestationType,
		SignCount:       int64(credential.Authenticator.SignCount),
	}, nil
}
