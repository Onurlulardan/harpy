package iam

import (
	"context"
	"errors"

	cryptography "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/cryptography"
	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
)

type recoveryAnchorVerifier struct{}

func (recoveryAnchorVerifier) Verify(
	ctx context.Context,
	wire []byte,
	expectation anchorapi.RegistrationExpectation,
) (anchorapi.RegistrationEvidence, error) {
	result, err := cryptography.VerifyRecoveryAnchorRegistration(ctx, wire,
		cryptography.RecoveryAnchorRegistrationExpectation{
			Challenge: expectation.Challenge, RPID: expectation.RPID,
			AllowedOrigins: expectation.AllowedOrigins})
	if err != nil {
		if errors.Is(err, cryptography.ErrRegistrationDenied) {
			return anchorapi.RegistrationEvidence{}, anchorapi.ErrDenied
		}
		return anchorapi.RegistrationEvidence{}, anchorapi.ErrInvalid
	}
	return anchorapi.RegistrationEvidence{
		CredentialID:  append([]byte(nil), result.CredentialID...),
		PublicKeyCOSE: append([]byte(nil), result.PublicKeyCOSE...),
		AAGUID:        result.AAGUID.String(), AttestationFormat: result.AttestationFormat,
		AttestationType: result.AttestationType, SignCount: result.SignCount}, nil
}

var _ anchorapi.RegistrationVerifier = recoveryAnchorVerifier{}
