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
}

func newIAMWorkerResources(
	maintenance *iammodule.MaintenanceJobRuntime,
	invitation *iammodule.InvitationDeliveryJobRuntime,
	lifecycle *iammodule.LifecycleDeliveryJobRuntime,
) (*iamWorkerResources, error) {
	if maintenance == nil || invitation == nil || lifecycle == nil {
		return nil, fmt.Errorf("IAM worker resources are incomplete")
	}
	return &iamWorkerResources{maintenance: maintenance,
		invitation: invitation, lifecycle: lifecycle}, nil
}

func (resources *iamWorkerResources) Close() error {
	if resources == nil {
		return nil
	}
	return errors.Join(resources.lifecycle.Close(),
		resources.invitation.Close(), resources.maintenance.Close())
}
