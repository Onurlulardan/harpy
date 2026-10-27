package operator

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops"
	systemresourcelayout "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func iamExecutor(owner tenancy.OperatorActivationExecutor) iam.InstallationActivationExecutor {
	return iam.InstallationActivationExecutor{QueryRow: func(
		ctx context.Context, statement string, arguments ...any,
	) iam.InstallationActivationRow {
		return owner.QueryRow(ctx, statement, arguments...)
	}}
}

func billingExecutor(
	owner tenancy.OperatorActivationExecutor,
) billing.InstallationActivationExecutor {
	return billing.InstallationActivationExecutor{QueryRow: func(
		ctx context.Context, statement string, arguments ...any,
	) billing.InstallationActivationRow {
		return owner.QueryRow(ctx, statement, arguments...)
	}}
}

func layoutExecutor(
	owner tenancy.OperatorActivationExecutor,
) systemresourcelayout.InstallationActivationExecutor {
	return systemresourcelayout.InstallationActivationExecutor{QueryRow: func(
		ctx context.Context, statement string, arguments ...any,
	) systemresourcelayout.InstallationActivationRow {
		return owner.QueryRow(ctx, statement, arguments...)
	}}
}

func platformOpsExecutor(
	owner tenancy.OperatorActivationExecutor,
) platformops.InstallationActivationExecutor {
	return platformops.InstallationActivationExecutor{QueryRow: func(
		ctx context.Context, statement string, arguments ...any,
	) platformops.InstallationActivationRow {
		return owner.QueryRow(ctx, statement, arguments...)
	}}
}

func authorityExecutor(
	owner tenancy.OperatorActivationExecutor,
) platformauthority.InstallationActivationExecutor {
	return platformauthority.InstallationActivationExecutor{QueryRow: func(
		ctx context.Context, statement string, arguments ...any,
	) platformauthority.InstallationActivationRow {
		return owner.QueryRow(ctx, statement, arguments...)
	}}
}

func asyncExecutor(
	owner tenancy.OperatorActivationExecutor,
) platformasync.InstallationActivationExecutor {
	return platformasync.InstallationActivationExecutor{QueryRow: func(
		ctx context.Context, statement string, arguments ...any,
	) platformasync.InstallationActivationRow {
		return owner.QueryRow(ctx, statement, arguments...)
	}}
}
