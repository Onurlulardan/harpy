package environmentoperator

import (
	"time"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func InstallationOffering(request OfferingRequest) (tenantapi.BaseOfferingPin, error) {
	if request.OperationID == uuid.Nil || len(request.Manifest.CanonicalBytes()) == 0 ||
		request.Manifest.SHA256() == "" || !validOfferingTime(request.CreatedAt) ||
		!knownInstallationMode(request.Manifest.DeploymentMode()) {
		return tenantapi.BaseOfferingPin{}, ErrInvalid
	}
	return tenantapi.BaseOfferingPin{
		VersionID:   tenantapi.UUID(stableID(request.Manifest, "installation-offering")),
		Revision:    1,
		SHA256:      request.Manifest.SHA256(),
		EffectiveAt: request.CreatedAt,
	}, nil
}

func knownInstallationMode(mode installationv1.DeploymentMode) bool {
	return mode == installationv1.DeploymentOnPremise || mode == installationv1.DeploymentSaaS
}

func validOfferingTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1_000 == 0
}
