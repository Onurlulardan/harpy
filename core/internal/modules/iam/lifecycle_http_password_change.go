package iam

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *lifecycleHTTPApplication) ChangePassword(
	ctx context.Context, claims accesstoken.Claims, input LifecyclePasswordChangeInput,
) error {
	defer clear(input.CurrentPassword)
	defer clear(input.NewPassword)
	if input.ExpectedRevision < 1 || !identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return ErrLifecycleHTTPInvalid
	}
	operationID := lifecycleOperationID(claims, "iam.password.change", input.IdempotencyKey)
	requestSHA, _ := lifecycleRequestSHA(struct{ Revision int64 }{input.ExpectedRevision})
	var stored lifecycle.PasswordCredential
	err := application.inSnapshot(ctx, claims, "read",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			var readErr error
			stored, readErr = application.runtime.service.PasswordCredentialLookup(scoped, scope,
				lifecycle.PasswordChangeCommand{
					Evidence: lifecycleEvidence(claims, permissionID, operationID, requestSHA),
					UserID:   claims.Subject(), ExpectedCredentialRevision: input.ExpectedRevision,
				})
			return readErr
		})
	if err != nil {
		return err
	}
	matched, err := verifyLifecyclePassword(input.CurrentPassword, stored)
	if err != nil || !matched {
		return ErrLifecycleHTTPInvalidCredential
	}
	credential, err := application.passwordCredential(ctx, input.NewPassword, nil)
	if err != nil {
		return err
	}
	return application.inCommand(ctx, claims, "iam.password.change", "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			_, executeErr := application.runtime.service.ChangePassword(scoped, scope,
				lifecycle.PasswordChangeCommand{
					Evidence: lifecycleEvidence(claims, permissionID, operationID, requestSHA),
					UserID:   claims.Subject(), ExpectedCredentialRevision: input.ExpectedRevision,
					Salt: credential.Salt(), Tag: credential.Tag(),
				})
			return executeErr
		})
}

func verifyLifecyclePassword(wire []byte, stored lifecycle.PasswordCredential) (bool, error) {
	secret, err := password.NewVerificationSecret(wire)
	if err != nil {
		return false, err
	}
	defer secret.Destroy()
	if stored.Algorithm == password.AlgorithmArgon2idV1 {
		credential, err := password.Argon2idCredentialFromStorage(stored.Salt, stored.Tag,
			uint32(stored.MemoryKiB), uint32(stored.Iterations), uint8(stored.Parallelism))
		if err != nil {
			return false, err
		}
		return password.Verify(secret, credential)
	}
	credential, err := password.LegacyBcryptCredentialFromStorage(
		stored.LegacyBcryptHash, stored.CreatedAt)
	if err != nil {
		return false, err
	}
	return password.VerifyLegacy(secret, credential, time.Now().UTC())
}
