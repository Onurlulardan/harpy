package iam

import (
	"context"
	"fmt"

	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	profilepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/profile"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type InvitationHTTPPlatform interface {
	ResolveActiveTenantIDBySlug(context.Context, string) (uuid.UUID, error)
	ResolveSnapshotDomainGenerations(
		context.Context, uuid.UUID, []string,
	) (map[string]int64, error)
}

type InvitationDeliverySealer interface {
	SealInvitationCredential(
		context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, []byte,
	) (deliveryartifactv1.Envelope, error)
}

type invitationHTTPApplication struct {
	pool        InvitationHTTPPlatform
	coordinator *platformtx.Coordinator
	runtime     *InvitationRuntime
	resolver    platformtx.CredentialResolverBinder
	keys        config.IAMOneTimeKeyring
	directory   identityDirectoryKey
	passwords   password.Policy
	hasher      password.Hasher
	sealer      InvitationDeliverySealer
	profiles    func([]byte, string, string, []byte) (invitationProfile, error)
}

type invitationProfile struct {
	schemaID uuid.UUID
	values   []byte
	digest   [32]byte
}

func NewInvitationHTTPApplication(
	pool InvitationHTTPPlatform,
	coordinator *platformtx.Coordinator,
	runtime *InvitationRuntime,
	settings config.IAMAPI,
	sealer InvitationDeliverySealer,
	commonPasswords IdentityPasswordRejector,
	compromisedPasswords IdentityPasswordRejector,
) (InvitationHTTPApplication, error) {
	if pool == nil || coordinator == nil || !runtime.Valid() || sealer == nil ||
		nilIdentityDependency(commonPasswords) || nilIdentityDependency(compromisedPasswords) {
		return nil, fmt.Errorf("IAM invitation HTTP dependencies are required")
	}
	resolver, err := newInvitationResolver()
	if err != nil {
		return nil, err
	}
	keys, err := settings.OneTimeTokenKeyring()
	if err != nil {
		return nil, fmt.Errorf("load IAM invitation keys: %w", err)
	}
	directory, err := identityDirectoryFromConfig(settings)
	if err != nil {
		return nil, err
	}
	policy, err := password.NewPolicy(commonPasswords, compromisedPasswords)
	if err != nil {
		clear(directory.material[:])
		return nil, err
	}
	return &invitationHTTPApplication{
		pool: pool, coordinator: coordinator, runtime: runtime, resolver: resolver,
		keys: keys, directory: directory, passwords: policy,
		hasher: password.NewHasher(), sealer: sealer,
		profiles: prepareInvitationProfile,
	}, nil
}

func prepareInvitationProfile(
	catalog []byte, firstName, lastName string, custom []byte,
) (invitationProfile, error) {
	document, err := profilepostgres.PrepareInvitationUserProfile(
		catalog, firstName, lastName, custom,
	)
	if err != nil {
		return invitationProfile{}, ErrInvitationHTTPInvalid
	}
	return invitationProfile{
		schemaID: uuid.UUID(document.SchemaID()),
		values:   document.Values(), digest: document.SHA256(),
	}, nil
}
