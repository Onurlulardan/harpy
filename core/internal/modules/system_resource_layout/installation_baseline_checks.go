package systemresourcelayout

import "fmt"

func checkInstallationBaselineActivation(valid bool, err error) error {
	if err != nil {
		return fmt.Errorf("assert system resource installation activation: %w", err)
	}
	if !valid {
		return fmt.Errorf("assert system resource installation activation: %w",
			ErrInstallationActivationUnverified)
	}
	return nil
}
