package lifecycle

import (
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	"github.com/google/uuid"
)

func scanInstallation(source row) (catalogdomain.Installation, error) {
	var installationID, packageUUID uuid.UUID
	var activeVersion *string
	var lastOperation *uuid.UUID
	var state string
	var referenceEpoch, rowVersion int64
	if err := source.Scan(&installationID, &packageUUID, &activeVersion, &state,
		&referenceEpoch, &rowVersion, &lastOperation); err != nil {
		return catalogdomain.Installation{}, err
	}
	id, idErr := packagecontract.ParseStableID(installationID.String())
	packageID, packageErr := packagecontract.ParsePackageID(packageUUID.String())
	if idErr != nil || packageErr != nil || referenceEpoch < 1 || rowVersion < 1 ||
		!validInstallationState(catalogdomain.InstallationState(state), activeVersion != nil) {
		return catalogdomain.Installation{}, unavailable()
	}
	value := catalogdomain.Installation{ID: id, PackageID: packageID,
		State: catalogdomain.InstallationState(state), ReferenceEpoch: uint64(referenceEpoch),
		RowVersion: uint64(rowVersion)}
	if activeVersion != nil {
		version, err := packagecontract.ParseVersion(*activeVersion)
		if err != nil {
			return catalogdomain.Installation{}, unavailable()
		}
		value.ActiveVersion = version
	}
	if lastOperation != nil {
		operationID, err := packagecontract.ParseStableID(lastOperation.String())
		if err != nil {
			return catalogdomain.Installation{}, unavailable()
		}
		value.LastOperationID = operationID
	}
	return value, nil
}

func validInstallationState(state catalogdomain.InstallationState, version bool) bool {
	switch state {
	case catalogdomain.InstallationInstalling, catalogdomain.InstallationFailed:
		return !version
	case catalogdomain.InstallationActive, catalogdomain.InstallationDisabled,
		catalogdomain.InstallationRetired, catalogdomain.InstallationPurgePending,
		catalogdomain.InstallationPurged:
		return version
	default:
		return false
	}
}
