package iam

import (
	"context"
	"fmt"
	"net/url"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	cryptography "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/cryptography"
	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type LifecycleHTTPPlatform interface {
	ResolveActiveTenantIDBySlug(context.Context, string) (uuid.UUID, error)
	ResolveSnapshotDomainGenerations(context.Context, uuid.UUID, []string) (map[string]int64, error)
	ResolveEnvironmentID(context.Context) (uuid.UUID, error)
}

type LifecycleDeliverySealer interface {
	SealLifecycleCredential(context.Context, uuid.UUID, uuid.UUID, uuid.UUID,
		uuid.UUID, uuid.UUID, lifecycledeliveryv1.Purpose, []byte,
	) (lifecycledeliveryv1.Envelope, error)
}

type LifecycleTOTPSealer interface {
	SealTOTP(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID,
		[32]byte) (TOTPEnvelope, error)
}

type lifecycleHTTPApplication struct {
	pool         LifecycleHTTPPlatform
	coordinator  *platformtx.Coordinator
	runtime      *LifecycleRuntime
	authorizer   authorization.Authorizer
	resolver     platformtx.CredentialResolverBinder
	oneTime      config.IAMOneTimeKeyring
	recovery     config.IAMRecoveryKeyring
	directory    identityDirectoryKey
	passwords    password.Policy
	hasher       password.Hasher
	delivery     LifecycleDeliverySealer
	totpSealer   LifecycleTOTPSealer
	totp         cryptography.TOTP
	origin, rpID string
}

func NewLifecycleHTTPApplication(
	pool LifecycleHTTPPlatform,
	coordinator *platformtx.Coordinator,
	runtime *LifecycleRuntime,
	settings config.IAMAPI,
	delivery LifecycleDeliverySealer,
	totpSealer LifecycleTOTPSealer,
	totpDecryptor TOTPDecryptor,
	common, compromised IdentityPasswordRejector,
) (LifecycleHTTPApplication, error) {
	if pool == nil || coordinator == nil || !runtime.Valid() || delivery == nil ||
		totpSealer == nil || totpDecryptor == nil || nilIdentityDependency(common) ||
		nilIdentityDependency(compromised) {
		return nil, fmt.Errorf("IAM lifecycle HTTP dependencies are required")
	}
	port, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build lifecycle authorizer: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(port)
	if err != nil {
		return nil, fmt.Errorf("build lifecycle policy: %w", err)
	}
	resolver, err := newInvitationResolver()
	if err != nil {
		return nil, err
	}
	oneTime, err := settings.OneTimeTokenKeyring()
	if err != nil {
		return nil, err
	}
	recovery, err := settings.RecoveryCodeKeyring()
	if err != nil {
		return nil, err
	}
	directory, err := identityDirectoryFromConfig(settings)
	if err != nil {
		return nil, err
	}
	policy, err := password.NewPolicy(common, compromised)
	if err != nil {
		return nil, err
	}
	profile, err := settings.HTTPProfile()
	if err != nil {
		return nil, err
	}
	origin := profile.AllowedOrigins()[0]
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Hostname() == "" {
		return nil, fmt.Errorf("IAM lifecycle origin is invalid")
	}
	totp, err := cryptography.NewTOTP(totpDecryptor)
	if err != nil {
		return nil, err
	}
	return &lifecycleHTTPApplication{
		pool: pool, coordinator: coordinator, runtime: runtime,
		authorizer: authorizer, resolver: resolver, oneTime: oneTime, recovery: recovery,
		directory: directory, passwords: policy, hasher: password.NewHasher(),
		delivery: delivery, totpSealer: totpSealer, totp: totp,
		origin: origin, rpID: parsed.Hostname(),
	}, nil
}

var _ lifecycle.OneTimeKeyring = config.IAMOneTimeKeyring{}
