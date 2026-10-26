package iam

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/gcpkms"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func OpenInvitationDeliveryJobRuntime(
	ctx context.Context,
	settings config.IAMInvitationSMTP,
	coordinator *platformtx.Coordinator,
) (*InvitationDeliveryJobRuntime, error) {
	client, err := openInvitationSMTP(ctx, settings)
	if err != nil {
		return nil, err
	}
	opener, err := gcpkms.OpenDeliveryOpener(ctx)
	if err != nil {
		return nil, fmt.Errorf("open IAM invitation delivery KMS: %w", err)
	}
	runtime, err := NewInvitationDeliveryJobRuntime(coordinator, opener, client)
	if err != nil {
		_ = opener.Close()
		return nil, fmt.Errorf("build IAM invitation delivery runtime: %w", err)
	}
	return runtime, nil
}
