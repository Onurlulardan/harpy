package api

import (
	"context"
	"fmt"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *iamAPIRuntime) buildSystemTransferHTTP(
	ctx context.Context,
	settings config.API,
	pool *postgres.Pool,
	coordinator *platformtx.Coordinator,
	cursorKeys config.CursorKeyring,
	catalogs platformasync.WriteCatalogs,
	report httpedge.ErrorSink,
) error {
	artifactSettings, err := settings.IAMSystemTransferArtifacts()
	if err != nil {
		return fmt.Errorf("load IAM system-transfer artifacts: %w", err)
	}
	artifacts, err := openIAMSystemTransferArtifacts(ctx, artifactSettings)
	if err != nil {
		return fmt.Errorf("build IAM system-transfer artifacts: %w", err)
	}
	application, err := iammodule.NewSystemTransferPlatformRuntime(
		pool, coordinator, runtime.identityApplication, runtime.restQueryRuntime,
		cursorKeys, catalogs, artifacts,
	)
	if err != nil {
		return fmt.Errorf("build IAM system-transfer application: %w", err)
	}
	access, err := iammodule.NewSystemTransferAccessRuntime(
		pool, coordinator, runtime.security, time.Now,
	)
	if err != nil {
		return fmt.Errorf("build IAM system-transfer access: %w", err)
	}
	handler, err := iammodule.NewSystemTransferHTTP(
		iammodule.SystemTransferHTTPDependencies{Access: access,
			Application: application,
			Report:      iammodule.SystemTransferHTTPErrorSink(report)},
	)
	if err != nil {
		return fmt.Errorf("build IAM system-transfer HTTP: %w", err)
	}
	runtime.systemTransferHTTP = handler
	runtime.routes = append(runtime.routes, handler.Routes())
	return nil
}
