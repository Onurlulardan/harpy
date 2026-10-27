package iam

import (
	"fmt"

	cryptography "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/cryptography"
	authpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func buildHTTPAuthenticator(
	security *HTTPSecurity,
	settings config.IAMAPI,
	decryptor cryptography.TOTPSecretDecryptor,
	passwordProvider cryptography.Password,
) (authentication.Authenticator, authentication.InitialPasswordAuthenticator, error) {
	store, err := authpostgres.NewStore()
	if err != nil {
		return authentication.Authenticator{}, authentication.InitialPasswordAuthenticator{}, err
	}
	refreshKeys, err := settings.RefreshTokenKeyring()
	if err != nil {
		return authentication.Authenticator{}, authentication.InitialPasswordAuthenticator{}, err
	}
	challengeKeys, err := settings.OneTimeTokenKeyring()
	if err != nil {
		return authentication.Authenticator{}, authentication.InitialPasswordAuthenticator{}, err
	}
	csrfKeys, err := settings.CSRFKeyring()
	if err != nil {
		return authentication.Authenticator{}, authentication.InitialPasswordAuthenticator{}, err
	}
	recoveryKeys, err := settings.RecoveryCodeKeyring()
	if err != nil {
		return authentication.Authenticator{}, authentication.InitialPasswordAuthenticator{}, err
	}
	access, err := cryptography.NewAccessIssuer(
		security.accessProfile, security.accessSigner,
	)
	if err != nil {
		return authentication.Authenticator{}, authentication.InitialPasswordAuthenticator{}, err
	}
	totp, err := cryptography.NewTOTP(decryptor)
	if err != nil {
		return authentication.Authenticator{}, authentication.InitialPasswordAuthenticator{}, err
	}
	authenticator, err := authentication.NewAuthenticator(
		store, passwordProvider, cryptography.NewRefreshGenerator(),
		refreshKeys, access, authentication.SecurityCryptography{
			Challenges:    cryptography.NewChallengeGenerator(),
			ChallengeKeys: challengeKeys, RecoveryKeys: recoveryKeys,
			WebAuthn: cryptography.NewWebAuthn(), TOTP: totp,
		}, authentication.SystemUUID, authentication.SystemClock,
	)
	if err != nil {
		return authentication.Authenticator{}, authentication.InitialPasswordAuthenticator{}, fmt.Errorf(
			"build IAM authenticator: %w", err,
		)
	}
	initialPort, ok := store.(authentication.InitialPasswordPort)
	if !ok {
		return authentication.Authenticator{}, authentication.InitialPasswordAuthenticator{},
			fmt.Errorf("build IAM initial password store: incompatible port")
	}
	initial, err := authentication.NewInitialPasswordAuthenticator(
		authenticator, initialPort, challengeKeys, csrfKeys,
	)
	if err != nil {
		return authentication.Authenticator{}, authentication.InitialPasswordAuthenticator{},
			fmt.Errorf("build IAM initial password authenticator: %w", err)
	}
	return authenticator, initial, nil
}
