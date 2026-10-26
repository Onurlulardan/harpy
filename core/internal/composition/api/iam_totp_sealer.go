package api

import (
	"context"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type iamTOTPSealerOpener func(context.Context, string) (*iammodule.TOTPSealer, error)

func openIAMTOTPSealer(ctx context.Context, settings config.IAMAPI,
	opener iamTOTPSealerOpener) (*iammodule.TOTPSealer, error) {
	if ctx == nil || ctx.Err() != nil || opener == nil {
		return nil, fmt.Errorf("IAM TOTP sealer dependencies are incomplete")
	}
	configured, err := settings.TOTPKMSKeyID()
	if err != nil {
		return nil, fmt.Errorf("load IAM TOTP KMS key ID: %w", err)
	}
	keyID, err := configured.Value()
	if err != nil {
		return nil, fmt.Errorf("read IAM TOTP KMS key ID: %w", err)
	}
	sealer, err := opener(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("open IAM TOTP sealer: %w", err)
	}
	return sealer, nil
}

func openGoogleIAMTOTPSealer(ctx context.Context, keyID string) (*iammodule.TOTPSealer, error) {
	return iammodule.OpenTOTPSealer(ctx, keyID)
}
