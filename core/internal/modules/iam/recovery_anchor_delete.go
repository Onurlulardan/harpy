package iam

import (
	"context"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
)

func (runtime *RecoveryAnchorRuntime) Delete(
	ctx context.Context,
	command anchorapi.DeleteCommand,
) (anchorapi.Operation, error) {
	if runtime == nil || runtime.inner == nil {
		return anchorapi.Operation{}, anchorapi.ErrUnavailable
	}
	return runtime.inner.ProposeDeletion(ctx, command)
}
