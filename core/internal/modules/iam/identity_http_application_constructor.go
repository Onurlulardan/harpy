package iam

import (
	"context"
	"fmt"
	"reflect"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type IdentityPasswordRejector interface {
	Blocked(context.Context, []byte) (bool, error)
}

type identityDirectoryKey struct {
	id       string
	material [32]byte
}

type identityHTTPApplication struct {
	pool        IdentityHTTPPlatform
	coordinator *platformtx.Coordinator
	identities  *IdentityRuntime
	queries     *queryapp.Service
	authorizer  authorization.Authorizer
	passwords   password.Policy
	hasher      password.Hasher
	directory   identityDirectoryKey
	changes     IdentityMutationSink
}

func NewIdentityHTTPApplication(
	pool IdentityHTTPPlatform,
	coordinator *platformtx.Coordinator,
	identities *IdentityRuntime,
	queries *IAMRESTQueryRuntime,
	settings config.IAMAPI,
	cursorKeys config.CursorKeyring,
	commonPasswords IdentityPasswordRejector,
	compromisedPasswords IdentityPasswordRejector,
	changes IdentityMutationSink,
) (IdentityHTTPApplication, error) {
	if pool == nil || coordinator == nil || !identities.Valid() || !queries.Valid() ||
		nilIdentityDependency(commonPasswords) || nilIdentityDependency(compromisedPasswords) ||
		nilIdentityDependency(changes) {
		return nil, fmt.Errorf("IAM system-resource application dependencies are required")
	}
	authorizationPort, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM system-resource authorizer: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(authorizationPort)
	if err != nil {
		return nil, fmt.Errorf("build IAM authorization policy: %w", err)
	}
	policy, err := password.NewPolicy(commonPasswords, compromisedPasswords)
	if err != nil {
		return nil, err
	}
	directory, err := identityDirectoryFromConfig(settings)
	if err != nil {
		return nil, err
	}
	queryKeys, err := identityQueryKeysFromConfig(cursorKeys)
	if err != nil {
		return nil, err
	}
	queryService, err := queries.newSystemResourceQuery(queryKeys)
	if err != nil {
		return nil, err
	}
	return &identityHTTPApplication{
		pool: pool, coordinator: coordinator, identities: identities,
		queries: queryService, authorizer: authorizer, passwords: policy,
		hasher: password.NewHasher(), directory: directory, changes: changes,
	}, nil
}

func nilIdentityDependency(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}
