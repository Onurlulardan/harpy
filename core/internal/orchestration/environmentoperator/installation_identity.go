package environmentoperator

import (
	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	"github.com/google/uuid"
)

type InstallationIdentity struct {
	OperationID uuid.UUID
	TenantID    uuid.UUID
	FirstUserID uuid.UUID
}

func DeriveInstallationIdentity(manifest installationv1.Manifest) InstallationIdentity {
	if len(manifest.CanonicalBytes()) == 0 || manifest.RequestSHA256() == "" {
		return InstallationIdentity{}
	}
	return InstallationIdentity{
		OperationID: stableID(manifest, "operation"),
		TenantID:    stableID(manifest, "primary-tenant"),
		FirstUserID: stableID(manifest, "first-user"),
	}
}

func (identity InstallationIdentity) Valid() bool {
	return identity.OperationID.Version() == uuid.Version(7) &&
		identity.TenantID.Version() == uuid.Version(7) &&
		identity.FirstUserID.Version() == uuid.Version(7)
}
