package api

import (
	"context"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type iamTOTPDecryptor = iammodule.TOTPDecryptor

type iamTOTPDecryptorOpener func(
	context.Context,
	string,
) (iamTOTPDecryptor, error)

func openIAMTOTPDecryptor(
	ctx context.Context,
	settings config.IAMAPI,
	opener iamTOTPDecryptorOpener,
) (iamTOTPDecryptor, error) {
	if ctx == nil || ctx.Err() != nil || settings.Validate() != nil || opener == nil {
		return nil, fmt.Errorf("IAM TOTP KMS dependencies are incomplete")
	}
	configured, err := settings.TOTPKMSKeyID()
	if err != nil {
		return nil, fmt.Errorf("load IAM TOTP KMS key ID: %w", err)
	}
	keyID, err := configured.Value()
	if err != nil {
		return nil, fmt.Errorf("read IAM TOTP KMS key ID: %w", err)
	}
	decryptor, err := opener(ctx, keyID)
	if err != nil || decryptor == nil {
		return nil, fmt.Errorf("open IAM TOTP KMS runtime: %w", err)
	}
	return decryptor, nil
}

func openGoogleIAMTOTPDecryptor(
	ctx context.Context,
	keyID string,
) (iamTOTPDecryptor, error) {
	return iammodule.OpenTOTPDecryptor(ctx, keyID)
}
