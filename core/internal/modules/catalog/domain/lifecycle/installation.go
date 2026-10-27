package lifecycle

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

type InstallationState string

const (
	InstallationInstalling   InstallationState = "INSTALLING"
	InstallationActive       InstallationState = "ACTIVE"
	InstallationDisabled     InstallationState = "DISABLED"
	InstallationFailed       InstallationState = "INSTALL_FAILED"
	InstallationRetired      InstallationState = "RETIRED"
	InstallationPurgePending InstallationState = "PURGE_PENDING"
	InstallationPurged       InstallationState = "PURGED"
)

type Installation struct {
	ID              packagecontract.StableID
	PackageID       packagecontract.PackageID
	ActiveVersion   packagecontract.Version
	State           InstallationState
	ReferenceEpoch  uint64
	RowVersion      uint64
	LastOperationID packagecontract.StableID
}

func NewInstalling(id packagecontract.StableID, packageID packagecontract.PackageID) (Installation, error) {
	if !id.Valid() || !packageID.Valid() {
		return Installation{}, ErrInvalid
	}
	return Installation{ID: id, PackageID: packageID,
		State: InstallationInstalling, ReferenceEpoch: 1, RowVersion: 1}, nil
}

func (value Installation) Retry() (Installation, error) {
	return value.transition(InstallationInstalling, InstallationFailed)
}

func (value Installation) Activate(version packagecontract.Version) (Installation, error) {
	if !version.Valid() || value.State != InstallationInstalling {
		return Installation{}, ErrTransition
	}
	value.State, value.ActiveVersion = InstallationActive, version
	value.RowVersion++
	return value, nil
}

func (value Installation) Upgrade(version packagecontract.Version) (Installation, error) {
	if !version.Valid() || !value.ActiveVersion.Valid() ||
		(value.State != InstallationActive && value.State != InstallationDisabled) ||
		version.Compare(value.ActiveVersion) <= 0 {
		return Installation{}, ErrTransition
	}
	value.ActiveVersion, value.RowVersion = version, value.RowVersion+1
	return value, nil
}

func (value Installation) Fail() (Installation, error) {
	return value.transition(InstallationFailed, InstallationInstalling)
}

func (value Installation) Disable() (Installation, error) {
	return value.transition(InstallationDisabled, InstallationActive)
}

func (value Installation) Enable() (Installation, error) {
	return value.transition(InstallationActive, InstallationDisabled)
}

func (value Installation) Retire() (Installation, error) {
	if value.State != InstallationActive && value.State != InstallationDisabled {
		return Installation{}, ErrTransition
	}
	value.State, value.ReferenceEpoch = InstallationRetired, value.ReferenceEpoch+1
	value.RowVersion++
	return value, nil
}

func (value Installation) Restore() (Installation, error) {
	return value.transition(InstallationActive, InstallationRetired)
}

func (value Installation) BeginPurge() (Installation, error) {
	if value.State != InstallationRetired {
		return Installation{}, ErrTransition
	}
	value.State, value.ReferenceEpoch = InstallationPurgePending, value.ReferenceEpoch+1
	value.RowVersion++
	return value, nil
}

func (value Installation) CancelPurge() (Installation, error) {
	if value.State != InstallationPurgePending {
		return Installation{}, ErrTransition
	}
	value.State, value.ReferenceEpoch = InstallationRetired, value.ReferenceEpoch+1
	value.RowVersion++
	return value, nil
}

func (value Installation) FinishPurge() (Installation, error) {
	return value.transition(InstallationPurged, InstallationPurgePending)
}

func (value Installation) transition(next InstallationState, allowed ...InstallationState) (Installation, error) {
	for _, state := range allowed {
		if value.State == state {
			value.State, value.RowVersion = next, value.RowVersion+1
			return value, nil
		}
	}
	return Installation{}, ErrTransition
}
