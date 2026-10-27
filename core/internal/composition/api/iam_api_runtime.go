package api

import (
	"context"
	"fmt"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func buildDeploymentIAMAPIRuntime(
	ctx context.Context, settings config.API, pool *postgres.Pool,
	catalogs platformasync.WriteCatalogs, report httpedge.ErrorSink,
	policy tenancymodule.DeploymentPolicy,
) (iamAPIRuntime, error) {
	environmentID, err := tenancymodule.ParseUUID(policy.EnvironmentID())
	if err != nil {
		return iamAPIRuntime{}, fmt.Errorf("parse deployment environment identity: %w", err)
	}
	return buildIAMAPIRuntime(ctx, settings, pool, catalogs, report, environmentID)
}

func buildIAMAPIRuntime(
	ctx context.Context,
	settings config.API,
	pool *postgres.Pool,
	catalogs platformasync.WriteCatalogs,
	report httpedge.ErrorSink,
	environmentID [16]byte,
) (iamAPIRuntime, error) {
	runtime := iamAPIRuntime{}
	owned := true
	defer func() {
		if owned {
			runtime.Close()
		}
	}()
	var err error
	runtime.rbacRuntime, err = iammodule.NewRBACRuntime(catalogs, time.Now)
	if err != nil {
		return iamAPIRuntime{}, fmt.Errorf("build IAM RBAC runtime: %w", err)
	}
	runtime.permissionRequests, err = iammodule.NewPermissionRequestRuntime(
		catalogs, time.Now,
	)
	if err != nil {
		return iamAPIRuntime{}, fmt.Errorf("build IAM permission requests: %w", err)
	}
	runtime.invitationRuntime, err = iammodule.NewInvitationRuntime(catalogs, time.Now)
	if err != nil {
		return iamAPIRuntime{}, fmt.Errorf("build IAM invitation runtime: %w", err)
	}
	runtime.lifecycleRuntime, err = iammodule.NewLifecycleRuntime(catalogs, time.Now)
	if err != nil {
		return iamAPIRuntime{}, fmt.Errorf("build IAM lifecycle runtime: %w", err)
	}
	runtime.identityRuntime, err = iammodule.NewIdentityRuntime()
	if err != nil {
		return iamAPIRuntime{}, fmt.Errorf("build IAM identity runtime: %w", err)
	}
	runtime.restQueryRuntime, err = iammodule.NewIAMRESTQueryRuntime()
	if err != nil {
		return iamAPIRuntime{}, fmt.Errorf("build IAM REST Query runtime: %w", err)
	}
	runtime.profileCommands, err = newIAMProfileCommands()
	if err != nil {
		return iamAPIRuntime{}, fmt.Errorf("build IAM profile commands: %w", err)
	}
	runtime.profileBackfills, err = iammodule.NewProfileBackfillPublicationRuntime(
		catalogs, time.Now,
	)
	if err != nil {
		return iamAPIRuntime{}, fmt.Errorf("build IAM profile backfill publication: %w", err)
	}
	runtime.profileRecoveries, err = iammodule.NewProfileBackfillRecoveryRuntime(
		catalogs, time.Now,
	)
	if err != nil {
		return iamAPIRuntime{}, fmt.Errorf("build IAM profile backfill recovery: %w", err)
	}
	identityChanges, err := iammodule.NewIdentityMutationSink(catalogs, time.Now)
	if err != nil {
		return iamAPIRuntime{}, fmt.Errorf("build IAM identity change sink: %w", err)
	}
	if err := runtime.buildSecurity(
		ctx, settings, pool, catalogs, identityChanges, report, environmentID,
	); err != nil {
		return iamAPIRuntime{}, err
	}
	owned = false
	return runtime, nil
}
