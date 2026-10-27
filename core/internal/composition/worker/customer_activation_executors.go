package worker

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	systemresourcelayout "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func customerIAMExecutor(owner tenancy.OperatorActivationExecutor) iam.InstallationActivationExecutor {
	return iam.InstallationActivationExecutor{QueryRow: func(ctx context.Context,
		statement string, arguments ...any,
	) iam.InstallationActivationRow {
		return owner.QueryRow(ctx, statement, arguments...)
	}}
}

func customerBillingExecutor(
	owner tenancy.OperatorActivationExecutor,
) billing.InstallationActivationExecutor {
	return billing.InstallationActivationExecutor{QueryRow: func(ctx context.Context,
		statement string, arguments ...any,
	) billing.InstallationActivationRow {
		return owner.QueryRow(ctx, statement, arguments...)
	}}
}

func customerLayoutExecutor(
	owner tenancy.OperatorActivationExecutor,
) systemresourcelayout.InstallationActivationExecutor {
	return systemresourcelayout.InstallationActivationExecutor{QueryRow: func(
		ctx context.Context, statement string, arguments ...any,
	) systemresourcelayout.InstallationActivationRow {
		return owner.QueryRow(ctx, statement, arguments...)
	}}
}

func customerAuthorityExecutor(
	owner tenancy.OperatorActivationExecutor,
) platformauthority.InstallationActivationExecutor {
	return platformauthority.InstallationActivationExecutor{QueryRow: func(
		ctx context.Context, statement string, arguments ...any,
	) platformauthority.InstallationActivationRow {
		return owner.QueryRow(ctx, statement, arguments...)
	}}
}

func customerAsyncExecutor(
	owner tenancy.OperatorActivationExecutor,
) platformasync.InstallationActivationExecutor {
	return platformasync.InstallationActivationExecutor{QueryRow: func(
		ctx context.Context, statement string, arguments ...any,
	) platformasync.InstallationActivationRow {
		return owner.QueryRow(ctx, statement, arguments...)
	}}
}
