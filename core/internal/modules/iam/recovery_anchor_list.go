package iam

import (
	"context"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
)

func (runtime *RecoveryAnchorRuntime) List(
	ctx context.Context,
	subject anchorapi.Subject,
) ([]anchorapi.Anchor, error) {
	if runtime == nil || runtime.inner == nil {
		return nil, anchorapi.ErrUnavailable
	}
	return runtime.inner.Anchors(ctx, subject)
}
