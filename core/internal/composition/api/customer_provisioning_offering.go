package api

import (
	"context"

	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

type customerProvisioningOffering struct{}

func (customerProvisioningOffering) ResolveSaaSFreeBase(
	context.Context,
) (tenancymodule.BaseOfferingPin, error) {
	pin := billingmodule.SaaSFreeBaseInstallationPin()
	if !pin.IsValid() {
		return tenancymodule.BaseOfferingPin{}, billingmodule.ErrSaaSFreeBaseUnavailable
	}
	return tenancymodule.BaseOfferingPin{VersionID: tenancymodule.UUID(pin.VersionID),
		Revision: pin.Revision, SHA256: pin.SHA256, EffectiveAt: pin.EffectiveAt}, nil
}
