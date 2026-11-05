package iam

import "fmt"

func checkInstallationActivationAssertion(valid bool, err error) error {
	if err != nil {
		return fmt.Errorf("assert IAM installation activation: %w", err)
	}
	if !valid {
		return fmt.Errorf("assert IAM installation activation: %w",
			ErrInstallationActivationUnverified)
	}
	return nil
}

func checkInstallationSecurityEpoch(ready bool, err error) error {
	if err != nil {
		return fmt.Errorf("initialize IAM installation security epoch: %w", err)
	}
	if !ready {
		return fmt.Errorf("initialize IAM installation security epoch: %w",
			ErrInstallationSecurityEpochNotReady)
	}
	return nil
}

func checkInstallationDirectoryRevision(revision int64, err error) error {
	if err != nil {
		return fmt.Errorf("initialize IAM installation directory: %w", err)
	}
	if revision != 1 {
		return fmt.Errorf("initialize IAM installation directory: revision=%d: %w",
			revision, ErrInstallationDirectoryNotInitialized)
	}
	return nil
}
