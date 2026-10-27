package iam

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	passwordenrollment "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/passwordenrollment"
)

type identityPreparedUserCredential struct {
	directory      identityfacade.DirectoryProjection
	password       passwordenrollment.Credential
	passwordSHA256 [32]byte
}

func (application *identityHTTPApplication) prepareUserCredential(
	ctx context.Context,
	input IdentityUserCreate,
) (identityPreparedUserCredential, error) {
	email, err := identifier.NormalizeEmailV1(input.Email)
	if err != nil {
		return identityPreparedUserCredential{}, ErrIdentityHTTPInvalid
	}
	secret, err := application.passwords.NewSecret(
		ctx, input.Password,
		[]string{email.Display(), email.Normalized(), input.FirstName, input.LastName},
	)
	if err != nil {
		return identityPreparedUserCredential{}, ErrIdentityHTTPInvalid
	}
	credential, err := application.hasher.Hash(secret)
	secret.Destroy()
	if err != nil {
		return identityPreparedUserCredential{}, ErrIdentityHTTPUnavailable
	}
	enrollment, err := passwordenrollment.NewCredential(
		credential.Salt(), credential.Tag(),
	)
	if err != nil {
		return identityPreparedUserCredential{}, ErrIdentityHTTPUnavailable
	}
	lookup := hmac.New(sha256.New, application.directory.material[:])
	_, _ = lookup.Write([]byte(email.Normalized()))
	lookupBytes := lookup.Sum(nil)
	var lookupHMAC [32]byte
	copy(lookupHMAC[:], lookupBytes)
	clear(lookupBytes)
	directory, err := identityfacade.NewDirectoryProjection(
		email.Display(), email.Normalized(), email.NormalizedSHA256(),
		application.directory.id, lookupHMAC,
	)
	if err != nil {
		return identityPreparedUserCredential{}, ErrIdentityHTTPUnavailable
	}
	return identityPreparedUserCredential{
		directory: directory, password: enrollment,
		passwordSHA256: sha256.Sum256(input.Password),
	}, nil
}
