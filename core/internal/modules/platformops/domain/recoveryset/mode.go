package recoveryset

import "errors"

var (
	ErrModeSelection = errors.New("recovery mode selection is invalid")
	ErrRestoreGate   = errors.New("database restore gate rejected the request")
	ErrManifest      = errors.New("recovery manifest is invalid")
)

type ModeSelection struct {
	Mode                 DeploymentMode
	OnPremiseStorage     bool
	SaaSProviderRegistry bool
}

func (selection ModeSelection) Validate() error {
	switch selection.Mode {
	case ModeOnPremise:
		if selection.OnPremiseStorage && !selection.SaaSProviderRegistry {
			return nil
		}
	case ModeSaaS:
		if !selection.OnPremiseStorage && selection.SaaSProviderRegistry {
			return nil
		}
	}
	return ErrModeSelection
}

func (selection ModeSelection) AllowsDatabaseRecovery() bool {
	return selection.Validate() == nil && selection.Mode == ModeOnPremise
}
