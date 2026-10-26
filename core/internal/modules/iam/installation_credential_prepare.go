package iam

import (
	"bytes"
	"encoding/base64"

	installationcredential "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/installationcredential"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
)

const installationPasswordBytes = 24

func PrepareInstallationCredential(
	presented []byte,
) (PreparedInstallationCredential, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(string(presented))
	if err != nil || len(decoded) != installationPasswordBytes ||
		!bytes.Equal([]byte(base64.RawURLEncoding.EncodeToString(decoded)), presented) {
		clear(decoded)
		return PreparedInstallationCredential{}, ErrInstallationCredentialInvalid
	}
	clear(decoded)
	secret, err := password.NewVerificationSecret(presented)
	if err != nil {
		return PreparedInstallationCredential{}, ErrInstallationCredentialInvalid
	}
	defer secret.Destroy()
	hashed, err := password.NewHasher().Hash(secret)
	if err != nil {
		return PreparedInstallationCredential{}, ErrInstallationCredentialInvalid
	}
	credential, err := installationcredential.NewCredential(hashed.Salt(), hashed.Tag())
	if err != nil {
		return PreparedInstallationCredential{}, ErrInstallationCredentialInvalid
	}
	return PreparedInstallationCredential{value: credential}, nil
}
