package worker

import (
	"context"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func buildIAMSystemTransferArtifacts(
	ctx context.Context,
	settings config.Worker,
) (iammodule.SystemTransferArtifacts, error) {
	artifactSettings, err := settings.IAMSystemTransferArtifacts()
	if err != nil {
		return nil, fmt.Errorf("load IAM system transfer artifacts: %w", err)
	}
	runtime, err := openIAMSystemTransferArtifacts(ctx, artifactSettings)
	if err != nil {
		return nil, fmt.Errorf("open IAM system transfer artifacts: %w", err)
	}
	return runtime, nil
}
