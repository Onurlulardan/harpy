package worker

import (
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func buildIAMMaintenance(
	settings config.IAMWorker,
) (*iammodule.MaintenanceJobRuntime, error) {
	keyring, err := settings.LoginDirectoryKeyring()
	if err != nil {
		return nil, fmt.Errorf("load IAM maintenance keyring: %w", err)
	}
	configured, err := keyring.Keys()
	if err != nil {
		return nil, fmt.Errorf("read IAM maintenance keyring: %w", err)
	}
	defer destroyConfiguredDirectoryKeys(configured)
	keys := make([]iammodule.MaintenanceDirectoryKey, 0, len(configured))
	for index := range configured {
		id, idErr := configured[index].ID()
		material, materialErr := configured[index].Material()
		active, activeErr := configured[index].Active()
		if idErr != nil || materialErr != nil || activeErr != nil {
			clear(material[:])
			return nil, fmt.Errorf("read IAM maintenance key: invalid keyring")
		}
		key, keyErr := iammodule.NewMaintenanceDirectoryKey(id, material, active)
		clear(material[:])
		if keyErr != nil {
			return nil, fmt.Errorf("build IAM maintenance key: %w", keyErr)
		}
		keys = append(keys, key)
	}
	runtime, err := iammodule.NewMaintenanceJobRuntime(keys)
	if err != nil {
		return nil, fmt.Errorf("build IAM maintenance runtime: %w", err)
	}
	return runtime, nil
}

func destroyConfiguredDirectoryKeys(keys []config.IAMDirectoryKey) {
	for index := range keys {
		keys[index].Destroy()
	}
}
