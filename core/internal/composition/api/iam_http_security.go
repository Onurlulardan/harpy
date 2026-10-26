package api

import (
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type iamHTTPSecurity = iammodule.HTTPSecurity

func buildIAMHTTPSecurity(settings config.API) (*iamHTTPSecurity, error) {
	iamConfig, err := settings.IAM()
	if err != nil {
		return nil, fmt.Errorf("load IAM API config: %w", err)
	}
	security, err := iammodule.NewHTTPSecurity(iamConfig)
	if err != nil {
		return nil, err
	}
	return security, nil
}
