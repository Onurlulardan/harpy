package billing

import "fmt"

func checkBillingInstallationActivation(stage string, valid bool, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w", stage, err)
	}
	if !valid {
		return fmt.Errorf("%s: %w", stage, ErrInstallationActivationUnverified)
	}
	return nil
}

func checkBillingInstallationUsage(meterCount int, replay bool, err error) error {
	if err != nil {
		return fmt.Errorf("open billing installation usage: %w", err)
	}
	if meterCount < 1 {
		return fmt.Errorf("open billing installation usage: meters=%d replay=%t: %w",
			meterCount, replay, ErrInstallationUsageNotOpened)
	}
	return nil
}
