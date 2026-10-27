package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *lifecycleHTTPApplication) CompletePasswordReset(
	ctx context.Context, input LifecyclePasswordResetCompleteInput,
) error {
	defer clear(input.Password)
	defer clear(input.Token)
	credentialInput := LifecycleCredentialInput{
		TenantSlug: input.TenantSlug,
		Token:      input.Token,
	}
	if err := application.validateCredential(
		ctx, credentialInput, lifecycle.PurposePasswordReset,
		"iam.password_reset.complete",
	); err != nil {
		return err
	}
	requestPayload := append([]byte(nil), input.Password...)
	defer clear(requestPayload)
	credential, err := application.passwordCredential(ctx, input.Password,
		[]string{input.TenantSlug})
	if err != nil {
		return err
	}
	return application.completeCredential(ctx, credentialInput,
		lifecycle.PurposePasswordReset, "iam.password_reset.complete", requestPayload,
		func(scoped context.Context, scope platformtx.CredentialTenantScope,
			_ lifecycle.CredentialRecord, proof lifecycle.CredentialProof,
			evidence lifecycle.CredentialEvidence) error {
			_, executeErr := application.runtime.service.CompletePasswordReset(scoped, scope,
				lifecycle.PasswordResetCompleteCommand{
					Evidence: evidence, Proof: proof,
					Salt: credential.Salt(), Tag: credential.Tag(),
				})
			return executeErr
		})
}

func (application *lifecycleHTTPApplication) passwordCredential(
	ctx context.Context, wire []byte, terms []string,
) (interface {
	Salt() [16]byte
	Tag() [32]byte
}, error) {
	secret, err := application.passwords.NewSecret(ctx, wire, terms)
	if err != nil {
		return nil, ErrLifecycleHTTPInvalid
	}
	defer secret.Destroy()
	credential, err := application.hasher.Hash(secret)
	if err != nil {
		return nil, ErrLifecycleHTTPUnavailable
	}
	return credential, nil
}
