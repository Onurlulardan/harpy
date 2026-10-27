package iam

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	passwordenrollment "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/passwordenrollment"
)

func (application *identityHTTPApplication) prepareSystemUserCredential(
	ctx context.Context,
	fieldValues publicresource.FieldValues,
	password []byte,
) (identityPreparedUserCredential, error) {
	raw, err := fieldValues.Values()
	if err != nil {
		return identityPreparedUserCredential{}, ErrIdentityHTTPInvalid
	}
	values := identitySystemResourceValues{
		resource: publicresource.Users, values: raw,
	}
	emailText, _, err := values.text(publicresource.UserEmailFieldID, true)
	if err != nil {
		return identityPreparedUserCredential{}, ErrIdentityHTTPInvalid
	}
	email, err := identifier.NormalizeEmailV1(emailText)
	if err != nil {
		return identityPreparedUserCredential{}, ErrIdentityHTTPInvalid
	}
	terms := append([]string{email.Display(), email.Normalized()}, values.passwordTerms()...)
	secret, err := application.passwords.NewSecret(ctx, password, terms)
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
		passwordSHA256: sha256.Sum256(password),
	}, nil
}
