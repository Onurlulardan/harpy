package api

import (
	"context"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *iamAPIRuntime) buildHTTP(
	ctx context.Context,
	settings config.API,
	pool *postgres.Pool,
	coordinator *platformtx.Coordinator,
	iamSettings config.IAMAPI,
	catalogs platformasync.WriteCatalogs,
	identityChanges iammodule.IdentityMutationSink,
	report httpedge.ErrorSink,
) error {
	var err error
	runtime.authApplication, err = iammodule.NewAuthHTTPApplication(
		pool, coordinator, runtime.security, iamSettings, runtime.totpDecryptor,
		runtime.passwordBlocklists.Common(), runtime.passwordBlocklists.Compromised(),
	)
	if err != nil {
		return fmt.Errorf("build IAM auth application: %w", err)
	}
	authHTTP, err := runtime.security.NewAuthHTTP(
		runtime.authApplication, iammodule.HTTPErrorSink(report),
	)
	if err != nil {
		return fmt.Errorf("build IAM auth HTTP: %w", err)
	}
	cursorKeys, err := settings.RestQueryCursorKeyring()
	if err != nil {
		return fmt.Errorf("load REST Query cursor keys: %w", err)
	}
	runtime.identityApplication, err = iammodule.NewIdentityHTTPApplication(
		pool, coordinator, runtime.identityRuntime, runtime.restQueryRuntime,
		iamSettings, cursorKeys, runtime.passwordBlocklists.Common(),
		runtime.passwordBlocklists.Compromised(), identityChanges,
	)
	if err != nil {
		return fmt.Errorf("build IAM identity application: %w", err)
	}
	if err = runtime.buildSystemResources(coordinator); err != nil {
		return err
	}
	identityHTTP, err := runtime.security.NewIdentityHTTP(
		runtime.identityApplication, iammodule.HTTPErrorSink(report),
	)
	if err != nil {
		return fmt.Errorf("build IAM identity HTTP: %w", err)
	}
	systemResourceHTTP, err := buildSystemResourceRecordHTTP(
		runtime.systemResources, runtime.systemLayoutAccess, report,
	)
	if err != nil {
		return err
	}
	identityRoutes, err := rebindIdentitySystemResourceRoutes(
		identityHTTP.Routes(), systemResourceHTTP.Routes(),
	)
	if err != nil {
		return fmt.Errorf("rebind IAM system resource HTTP: %w", err)
	}
	runtime.postRestore, err = iammodule.NewPostRestoreRuntime(
		pool, iamSettings, runtime.passwordBlocklists.Common(),
		runtime.passwordBlocklists.Compromised(), authHTTP, iammodule.HTTPErrorSink(report),
	)
	if err != nil {
		return fmt.Errorf("build IAM post-restore runtime: %w", err)
	}
	runtime.routes = [][]httpedge.Route{
		authHTTP.Routes(), identityRoutes, systemResourceHTTP.Routes(),
		runtime.postRestore.Routes(),
	}
	if err := runtime.buildSystemTransferHTTP(ctx, settings, pool, coordinator,
		cursorKeys, catalogs, report); err != nil {
		return err
	}
	if err := runtime.buildProfileHTTP(
		pool, coordinator, identityChanges, report,
	); err != nil {
		return err
	}
	if err := runtime.buildInvitationHTTP(
		pool, coordinator, iamSettings, report,
	); err != nil {
		return err
	}
	if err := runtime.buildPermissionRequestHTTP(pool, coordinator, report); err != nil {
		return err
	}
	if err := runtime.buildLifecycleHTTP(pool, coordinator, iamSettings, report); err != nil {
		return err
	}
	if err := runtime.buildRBACHTTP(pool, coordinator, report); err != nil {
		return err
	}
	return nil
}
