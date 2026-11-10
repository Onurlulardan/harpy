package api

import (
	"context"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type closedProfileReferences struct{}

func newProfileReferences() iammodule.ProfileReferencePort {
	return closedProfileReferences{}
}

func newIAMProfileCommands() (*iammodule.ProfileCommands, error) {
	return iammodule.NewProfileCommands(newProfileReferences())
}

func (closedProfileReferences) AssertRetirable(
	context.Context,
	platformtx.TenantScope,
	iammodule.ProfileRetirementTarget,
) (iammodule.ProfileReferenceEvidence, error) {
	return iammodule.ProfileReferenceEvidence{}, fmt.Errorf("profile schema retirement is closed")
}
