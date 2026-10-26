package worker

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildIAMDelivery(ctx context.Context, settings config.IAMWorker,
	coordinator *platformtx.Coordinator,
) (*iam.InvitationDeliveryJobRuntime, *iam.LifecycleDeliveryJobRuntime, error) {
	smtpConfig, err := settings.InvitationSMTP()
	if err != nil {
		return nil, nil, fmt.Errorf("load IAM SMTP config: %w", err)
	}
	return iam.OpenDeliveryJobRuntimes(ctx, smtpConfig, coordinator)
}
