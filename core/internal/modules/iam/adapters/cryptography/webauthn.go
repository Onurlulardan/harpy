package cryptography

import (
	"bytes"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/go-webauthn/webauthn/protocol"
)

type WebAuthn struct{}

func NewWebAuthn() WebAuthn { return WebAuthn{} }

func (WebAuthn) Verify(
	assertion authentication.WebAuthnAssertion,
	expectation authentication.WebAuthnExpectation,
) (authentication.WebAuthnResult, error) {
	concrete, ok := assertion.(*WebAuthnAssertion)
	if !ok || concrete == nil || len(expectation.Challenge) != 43 ||
		len(expectation.CredentialID) == 0 || len(expectation.PublicKeyCOSE) == 0 ||
		expectation.RPID == "" || expectation.Origin == "" {
		return authentication.WebAuthnResult{}, authentication.ErrAuthenticationContract
	}
	wire, err := concrete.Bytes()
	if err != nil {
		return authentication.WebAuthnResult{}, err
	}
	defer clear(wire)
	parsed, err := protocol.ParseCredentialRequestResponseBytes(wire)
	if err != nil || parsed == nil ||
		!bytes.Equal(parsed.RawID, expectation.CredentialID) ||
		(len(parsed.Response.UserHandle) != 0 &&
			!bytes.Equal(parsed.Response.UserHandle, expectation.UserHandle)) {
		return authentication.WebAuthnResult{}, authentication.ErrInvalidCredentials
	}
	err = parsed.Verify(
		string(expectation.Challenge), expectation.RPID, "",
		[]string{expectation.Origin}, []string{expectation.Origin},
		protocol.TopOriginExplicitVerificationMode, false,
		expectation.UserVerificationRequired, true, expectation.PublicKeyCOSE,
	)
	if err != nil {
		return authentication.WebAuthnResult{}, authentication.ErrInvalidCredentials
	}
	flags := parsed.Response.AuthenticatorData.Flags
	return authentication.WebAuthnResult{
		SignCount:      parsed.Response.AuthenticatorData.Counter,
		BackupEligible: flags.HasBackupEligible(), BackupState: flags.HasBackupState(),
		UserVerified: flags.HasUserVerified(),
	}, nil
}
