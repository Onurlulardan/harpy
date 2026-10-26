package platformops

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type httpRecoveryAnchorFixture struct{}

func (*httpRecoveryAnchorFixture) ListRecoveryAnchors(
	context.Context, platformopsapi.RecoveryAnchorSubject,
) ([]platformopsapi.RecoveryAnchor, error) {
	return []platformopsapi.RecoveryAnchor{}, nil
}

func (*httpRecoveryAnchorFixture) CreateRecoveryAnchorOptions(
	context.Context, platformopsapi.RecoveryAnchorOptionsCommand,
) (platformopsapi.RecoveryAnchorRegistrationOptions, error) {
	return platformopsapi.RecoveryAnchorRegistrationOptions{}, nil
}

func (*httpRecoveryAnchorFixture) RegisterRecoveryAnchor(
	context.Context, platformopsapi.RecoveryAnchorRegisterCommand,
) (platformopsapi.RecoveryAnchorOperation, error) {
	return platformopsapi.RecoveryAnchorOperation{}, nil
}

func (*httpRecoveryAnchorFixture) DeleteRecoveryAnchor(
	context.Context, platformopsapi.RecoveryAnchorDeleteCommand,
) (platformopsapi.RecoveryAnchorOperation, error) {
	return platformopsapi.RecoveryAnchorOperation{}, nil
}
