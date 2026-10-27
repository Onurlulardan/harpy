package worker

import (
	"errors"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
)

type iamWorkerResources struct {
	maintenance *iammodule.MaintenanceJobRuntime
	invitation  *iammodule.InvitationDeliveryJobRuntime
	lifecycle   *iammodule.LifecycleDeliveryJobRuntime
	imports     *iammodule.SystemTransferImportJobRuntime
	exports     *iammodule.SystemTransferExportJobRuntime
	cleanup     *iammodule.SystemTransferImportCleanupRuntime
}

func newIAMWorkerResources(
	maintenance *iammodule.MaintenanceJobRuntime,
	invitation *iammodule.InvitationDeliveryJobRuntime,
	lifecycle *iammodule.LifecycleDeliveryJobRuntime,
	imports *iammodule.SystemTransferImportJobRuntime,
	exports *iammodule.SystemTransferExportJobRuntime,
	cleanup *iammodule.SystemTransferImportCleanupRuntime,
) (*iamWorkerResources, error) {
	if maintenance == nil || invitation == nil || lifecycle == nil || imports == nil ||
		exports == nil || cleanup == nil {
		return nil, fmt.Errorf("IAM worker resources are incomplete")
	}
	return &iamWorkerResources{maintenance: maintenance,
		invitation: invitation, lifecycle: lifecycle, imports: imports,
		exports: exports, cleanup: cleanup}, nil
}

func (resources *iamWorkerResources) Close() error {
	if resources == nil {
		return nil
	}
	return errors.Join(resources.exports.Close(), resources.imports.Close(),
		resources.lifecycle.Close(),
		resources.invitation.Close(), resources.maintenance.Close())
}
