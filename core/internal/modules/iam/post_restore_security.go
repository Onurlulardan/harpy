package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/cryptography"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
)

type postRestoreChallengeGenerator struct{}

func (postRestoreChallengeGenerator) Generate() (postrestore.IssuedChallenge, error) {
	return cryptography.NewChallengeGenerator().Generate()
}

type postRestoreAssertionVerifier struct{}

func (postRestoreAssertionVerifier) Verify(
	document []byte,
	expectation postrestore.AssertionExpectation,
) (postrestore.AssertionResult, error) {
	assertion, err := cryptography.NewWebAuthnAssertion(document)
	if err != nil {
		return postrestore.AssertionResult{}, err
	}
	defer assertion.Destroy()
	result, err := cryptography.NewWebAuthn().Verify(
		assertion,
		authentication.WebAuthnExpectation{
			Challenge: expectation.Challenge, CredentialID: expectation.CredentialID,
			PublicKeyCOSE: expectation.PublicKeyCOSE, RPID: expectation.RPID,
			Origin: expectation.Origin, UserHandle: expectation.UserHandle,
			UserVerificationRequired: expectation.UserVerificationRequired,
		},
	)
	if err != nil {
		return postrestore.AssertionResult{}, err
	}
	return postrestore.AssertionResult{
		SignCount: result.SignCount, BackupEligible: result.BackupEligible,
		BackupState: result.BackupState, UserVerified: result.UserVerified,
	}, nil
}

type postRestorePasswordIssuer struct {
	policy password.Policy
}

func (issuer postRestorePasswordIssuer) Issue(
	ctx context.Context,
	candidate []byte,
	contextTerms []string,
) (postrestore.PasswordCredential, error) {
	secret, err := issuer.policy.NewSecret(ctx, candidate, contextTerms)
	if err != nil {
		return postrestore.PasswordCredential{}, err
	}
	defer secret.Destroy()
	credential, err := password.NewHasher().Hash(secret)
	if err != nil {
		return postrestore.PasswordCredential{}, err
	}
	return postrestore.PasswordCredential{
		Salt: credential.Salt(), Tag: credential.Tag(),
		MemoryKiB: credential.MemoryKiB(), Iterations: credential.Iterations(),
		Parallelism: credential.Parallelism(),
	}, nil
}
