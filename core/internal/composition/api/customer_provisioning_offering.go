package api

import (
	"context"

	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

type customerProvisioningOffering struct {
	reader billingmodule.SaaSFreeBaseReader
}

func (resolver customerProvisioningOffering) ResolveSaaSFreeBase(
	ctx context.Context,
) (tenancymodule.BaseOfferingPin, error) {
	pin, err := billingmodule.ReadSaaSFreeBase(ctx, resolver.reader)
	if err != nil {
		return tenancymodule.BaseOfferingPin{}, err
	}
	return tenancymodule.BaseOfferingPin{VersionID: tenancymodule.UUID(pin.VersionID),
		Revision: pin.Revision, SHA256: pin.SHA256, EffectiveAt: pin.EffectiveAt}, nil
}
