package iam

import (
	"context"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
)

func (runtime *RecoveryAnchorRuntime) CreateOptions(
	ctx context.Context,
	command anchorapi.OptionsCommand,
) (anchorapi.RegistrationOptions, error) {
	if runtime == nil || runtime.inner == nil {
		return anchorapi.RegistrationOptions{}, anchorapi.ErrUnavailable
	}
	return runtime.inner.Options(ctx, command)
}
