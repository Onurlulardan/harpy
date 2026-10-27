package iam

import (
	"context"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
)

func (runtime *RecoveryAnchorRuntime) Register(
	ctx context.Context,
	command anchorapi.RegisterCommand,
) (anchorapi.Operation, error) {
	if runtime == nil || runtime.inner == nil {
		return anchorapi.Operation{}, anchorapi.ErrUnavailable
	}
	return runtime.inner.ProposeRegistration(ctx, command)
}
