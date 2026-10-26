package iam

import (
	"fmt"
	"reflect"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	profilepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/profile"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type profileHTTPApplication struct {
	pool        IdentityHTTPPlatform
	coordinator *platformtx.Coordinator
	commands    *profileapp.Commands
	selfWriter  profilepostgres.SelfWriter
	backfills   *ProfileBackfillPublicationRuntime
	recoveries  *ProfileBackfillRecoveryRuntime
	authorizer  authorization.Authorizer
	changes     IdentityMutationSink
}

func NewProfileHTTPApplication(
	pool IdentityHTTPPlatform,
	coordinator *platformtx.Coordinator,
	commands *profileapp.Commands,
	backfills *ProfileBackfillPublicationRuntime,
	recoveries *ProfileBackfillRecoveryRuntime,
	changes IdentityMutationSink,
) (ProfileHTTPApplication, error) {
	if pool == nil || coordinator == nil || commands == nil || backfills == nil ||
		recoveries == nil || nilProfileHTTPDependency(changes) {
		return nil, fmt.Errorf("IAM profile application dependencies are required")
	}
	port, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM profile authorizer: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(port)
	if err != nil {
		return nil, fmt.Errorf("build IAM profile authorization policy: %w", err)
	}
	selfWriter, err := profilepostgres.NewSelfWriter()
	if err != nil {
		return nil, fmt.Errorf("build IAM self profile writer: %w", err)
	}
	return &profileHTTPApplication{
		pool: pool, coordinator: coordinator, commands: commands, backfills: backfills,
		recoveries: recoveries,
		selfWriter: selfWriter, authorizer: authorizer, changes: changes,
	}, nil
}

func nilProfileHTTPDependency(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}
