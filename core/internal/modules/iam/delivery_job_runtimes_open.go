package iam

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/gcpkms"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func OpenDeliveryJobRuntimes(ctx context.Context,
	settings config.IAMInvitationSMTP, coordinator *platformtx.Coordinator,
) (*InvitationDeliveryJobRuntime, *LifecycleDeliveryJobRuntime, error) {
	invitationSMTP, err := openInvitationSMTP(ctx, settings)
	if err != nil {
		return nil, nil, err
	}
	lifecycleSMTP, err := openLifecycleSMTP(ctx, settings)
	if err != nil {
		return nil, nil, err
	}
	opener, err := gcpkms.OpenDeliveryOpener(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("open IAM delivery KMS: %w", err)
	}
	invitation, err := NewInvitationDeliveryJobRuntime(coordinator, opener, invitationSMTP)
	if err != nil {
		_ = opener.Close()
		return nil, nil, fmt.Errorf("build IAM invitation delivery runtime: %w", err)
	}
	lifecycle, err := NewLifecycleDeliveryJobRuntime(coordinator, opener, lifecycleSMTP)
	if err != nil {
		_ = invitation.Close()
		return nil, nil, fmt.Errorf("build IAM lifecycle delivery runtime: %w", err)
	}
	return invitation, lifecycle, nil
}
