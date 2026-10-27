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

func (runtime *iamAPIRuntime) buildSecurity(
	ctx context.Context,
	settings config.API,
	pool *postgres.Pool,
	catalogs platformasync.WriteCatalogs,
	identityChanges iammodule.IdentityMutationSink,
	report httpedge.ErrorSink,
	environmentID [16]byte,
) error {
	var err error
	runtime.security, err = buildIAMHTTPSecurity(settings)
	if err != nil {
		return err
	}
	iamSettings, err := settings.IAM()
	if err != nil {
		return fmt.Errorf("load IAM API settings: %w", err)
	}
	blocklistSettings, err := iamSettings.PasswordBlocklists()
	if err != nil {
		return fmt.Errorf("load IAM password blocklists: %w", err)
	}
	runtime.passwordBlocklists, err = iammodule.OpenPasswordBlocklists(blocklistSettings)
	if err != nil {
		return fmt.Errorf("open IAM password blocklists: %w", err)
	}
	runtime.totpDecryptor, err = openIAMTOTPDecryptor(
		ctx, iamSettings, openGoogleIAMTOTPDecryptor,
	)
	if err != nil {
		return err
	}
	runtime.totpSealer, err = openIAMTOTPSealer(
		ctx, iamSettings, openGoogleIAMTOTPSealer,
	)
	if err != nil {
		return err
	}
	runtime.deliverySealer, err = iammodule.OpenDeliverySealer(ctx, iamSettings)
	if err != nil {
		return fmt.Errorf("open IAM invitation delivery sealer: %w", err)
	}
	coordinator, err := platformtx.NewCoordinator(pool)
	if err != nil {
		return fmt.Errorf("build tenant coordinator: %w", err)
	}
	runtime.coordinator = coordinator
	runtime.identityChanges = identityChanges
	if err := runtime.buildHTTPAccess(pool, coordinator, catalogs, environmentID); err != nil {
		return err
	}
	if err := runtime.buildHTTP(ctx, settings, pool, coordinator, iamSettings,
		catalogs, identityChanges, report); err != nil {
		return err
	}
	return nil
}

func (runtime *iamAPIRuntime) buildActorBridge(settings config.API) error {
	iamSettings, err := settings.IAM()
	if err != nil {
		return fmt.Errorf("load IAM actor bridge settings: %w", err)
	}
	runtime.actorBridge, err = iammodule.NewActorBridgeRuntime(
		iamSettings, runtime.coordinator,
	)
	if err != nil {
		return fmt.Errorf("build IAM V1 actor bridge: %w", err)
	}
	return nil
}
