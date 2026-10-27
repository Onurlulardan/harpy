package iam

import (
	"context"
	"fmt"

	cryptography "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/cryptography"
	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	platformiampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/platformiam"
	sessionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/sessionview"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
	platformiam "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/platformiam"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/sessionview"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type AuthHTTPPlatform interface {
	ResolveActiveTenantIDBySlug(context.Context, string) (uuid.UUID, error)
	ResolveSnapshotDomainGenerations(
		context.Context, uuid.UUID, []string,
	) (map[string]int64, error)
	IAMThrottleBlocked(
		context.Context, string, uuid.UUID, string, string, string,
		[32]byte, int64, int64,
	) (bool, error)
	IAMThrottleRecordFailure(
		context.Context, string, uuid.UUID, string, string, string,
		[32]byte, int64, int64,
	) (bool, error)
}

type authHTTPApplication struct {
	pool          AuthHTTPPlatform
	coordinator   *platformtx.Coordinator
	resolver      platformtx.CredentialResolverBinder
	authenticator authentication.Authenticator
	initial       authentication.InitialPasswordAuthenticator
	passwords     password.Policy
	password      cryptography.Password
	dummyPassword authentication.Argon2idCredential
	directory     config.IAMDirectoryKeyring
	network       config.IAMSessionNetworkKeyring
	throttle      config.IAMThrottleKeyring
	sessions      *sessionview.Reader
	platformIAM   *platformiam.Service
}

func NewAuthHTTPApplication(
	pool AuthHTTPPlatform,
	coordinator *platformtx.Coordinator,
	security *HTTPSecurity,
	settings config.IAMAPI,
	decryptor cryptography.TOTPSecretDecryptor,
	common, compromised IdentityPasswordRejector,
) (AuthHTTPApplication, error) {
	if pool == nil || coordinator == nil || security == nil ||
		!security.Valid() || decryptor == nil || nilIdentityDependency(common) ||
		nilIdentityDependency(compromised) {
		return nil, fmt.Errorf("IAM auth HTTP runtime dependencies are required")
	}
	passwordProvider, dummyPassword, err := newHTTPLoginEqualizer()
	if err != nil {
		return nil, err
	}
	authenticator, initial, err := buildHTTPAuthenticator(
		security, settings, decryptor, passwordProvider,
	)
	if err != nil {
		return nil, err
	}
	policy, err := password.NewPolicy(common, compromised)
	if err != nil {
		return nil, err
	}
	resolver, err := iampostgres.NewCredentialTargetResolverBinder()
	if err != nil {
		return nil, fmt.Errorf("build IAM credential target resolver: %w", err)
	}
	directory, err := settings.LoginDirectoryKeyring()
	if err != nil {
		return nil, fmt.Errorf("load IAM login directory keys: %w", err)
	}
	network, err := settings.SessionNetworkKeyring()
	if err != nil {
		return nil, fmt.Errorf("load IAM session network keys: %w", err)
	}
	throttle, err := settings.AuthThrottleKeyring()
	if err != nil {
		return nil, fmt.Errorf("load IAM auth throttle keys: %w", err)
	}
	sessionStore, err := sessionpostgres.NewStore()
	if err != nil {
		return nil, fmt.Errorf("build IAM session view store: %w", err)
	}
	sessions, err := sessionview.NewReader(sessionStore)
	if err != nil {
		return nil, fmt.Errorf("build IAM session view reader: %w", err)
	}
	platformStore, err := platformiampostgres.NewStore()
	if err != nil {
		return nil, fmt.Errorf("build IAM platform authority store: %w", err)
	}
	platformAuthority, err := platformiam.NewService(platformStore, iamPlatformClock{})
	if err != nil {
		return nil, fmt.Errorf("build IAM platform authority service: %w", err)
	}
	return &authHTTPApplication{
		pool: pool, coordinator: coordinator, resolver: resolver,
		authenticator: authenticator, initial: initial, passwords: policy,
		password:      passwordProvider,
		dummyPassword: dummyPassword, directory: directory,
		network: network, throttle: throttle, sessions: sessions,
		platformIAM: platformAuthority,
	}, nil
}
