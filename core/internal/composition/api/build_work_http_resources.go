package api

import (
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	workmodule "github.com/Onurlulardan/harpy/core/internal/modules/work"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func buildWorkHTTPResources(settings config.API) (
	workmodule.PlatformHTTPRuntimeDependencies, error,
) {
	iamResources, err := iammodule.NewWorkResourceRuntime()
	if err != nil {
		return workmodule.PlatformHTTPRuntimeDependencies{},
			fmt.Errorf("build Work HTTP resource runtime: %w", err)
	}
	projection, err := workmodule.NewIAMResourceAuthority(iamResources)
	if err != nil {
		return workmodule.PlatformHTTPRuntimeDependencies{},
			fmt.Errorf("build Work HTTP resource authority: %w", err)
	}
	taskReferences, err := workmodule.NewClosedTaskReferences()
	if err != nil {
		return workmodule.PlatformHTTPRuntimeDependencies{},
			fmt.Errorf("build Work HTTP task references: %w", err)
	}
	taskResources, err := workmodule.NewClosedTaskResources()
	if err != nil {
		return workmodule.PlatformHTTPRuntimeDependencies{},
			fmt.Errorf("build Work HTTP task resources: %w", err)
	}
	cursorKeys, err := workHTTPCursorKeys(settings)
	if err != nil {
		return workmodule.PlatformHTTPRuntimeDependencies{}, err
	}
	return workmodule.PlatformHTTPRuntimeDependencies{
		CursorKeys: cursorKeys, ProjectionResources: projection,
		TaskReferences: taskReferences, TaskResources: taskResources,
	}, nil
}

func workHTTPCursorKeys(settings config.API) (workmodule.RESTQueryCursorKeys, error) {
	keyring, err := settings.RestQueryCursorKeyring()
	if err != nil {
		return workmodule.RESTQueryCursorKeys{}, fmt.Errorf("load Work HTTP cursor keys: %w", err)
	}
	current, err := keyring.Current()
	if err != nil {
		return workmodule.RESTQueryCursorKeys{}, fmt.Errorf("load Work current cursor key: %w", err)
	}
	previous, err := keyring.Previous()
	if err != nil {
		return workmodule.RESTQueryCursorKeys{}, fmt.Errorf("load Work previous cursor key: %w", err)
	}
	currentID, currentIDErr := current.KID()
	currentKey, currentKeyErr := current.AES256Key()
	previousID, previousIDErr := previous.KID()
	previousKey, previousKeyErr := previous.AES256Key()
	if currentIDErr != nil || currentKeyErr != nil ||
		previousIDErr != nil || previousKeyErr != nil {
		clear(currentKey[:])
		clear(previousKey[:])
		return workmodule.RESTQueryCursorKeys{}, fmt.Errorf("load Work REST Query cursor keys")
	}
	keys, err := workmodule.NewRESTQueryCursorKeys(
		currentID, currentKey, previousID, previousKey,
	)
	clear(currentKey[:])
	clear(previousKey[:])
	if err != nil {
		return workmodule.RESTQueryCursorKeys{}, fmt.Errorf("build Work cursor keys: %w", err)
	}
	return keys, nil
}
