package api

import (
	"fmt"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	layoutmodule "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout"
	systemresourcelayouts "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresourcelayouts"
	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func newIAMProfileCommands() (*iammodule.ProfileCommands, error) {
	return iammodule.NewProfileCommands(
		newProfileReferences(layoutmodule.NewReferences()),
	)
}

func (runtime *iamAPIRuntime) buildSystemResources(
	coordinator *platformtx.Coordinator,
) error {
	layout, err := layoutmodule.NewPresentationRuntime(time.Now)
	if err != nil {
		return fmt.Errorf("build system-resource presentation runtime: %w", err)
	}
	runtime.systemResources, err = systemresources.NewApplication(
		coordinator, runtime.identityApplication.SystemResources(), layout,
	)
	if err != nil {
		return fmt.Errorf("build system-resource orchestrator: %w", err)
	}
	admin, err := layoutmodule.NewProductionLayoutRuntime(time.Now)
	if err != nil {
		return fmt.Errorf("build system-resource layout runtime: %w", err)
	}
	runtime.systemLayouts, err = systemresourcelayouts.NewApplication(
		coordinator, runtime.identityApplication.SystemResources(), admin,
	)
	if err != nil {
		return fmt.Errorf("build system-resource layout orchestrator: %w", err)
	}
	return nil
}
