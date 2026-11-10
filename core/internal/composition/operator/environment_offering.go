package operator

import (
	"context"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/orchestration/environmentoperator"
)

type EnvironmentOffering struct{}

func NewEnvironmentOffering() *EnvironmentOffering {
	return &EnvironmentOffering{}
}

func (resolver *EnvironmentOffering) Resolve(
	ctx context.Context,
	request environmentoperator.OfferingRequest,
) (tenancymodule.BaseOfferingPin, error) {
	if resolver == nil || ctx == nil || ctx.Err() != nil {
		return tenancymodule.BaseOfferingPin{}, environmentoperator.ErrInvalid
	}
	return environmentoperator.InstallationOffering(request)
}

var _ environmentoperator.Offering = (*EnvironmentOffering)(nil)
