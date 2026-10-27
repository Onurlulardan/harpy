package tenancy

import (
	"context"
	"fmt"

	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
	"github.com/google/uuid"
)

type CustomerProvisioningApplication interface {
	Create(context.Context, provisioning.Command) (provisioning.Operation, error)
	Read(context.Context, provisioning.ReadQuery) (provisioning.Operation, error)
	ReissueInitialAdminSetup(context.Context, provisioning.ReissueCommand) (
		provisioning.Operation, error,
	)
}

type CustomerProvisioningHTTP struct {
	application CustomerProvisioningApplication
	report      func(error)
}

func NewCustomerProvisioningHTTP(
	application CustomerProvisioningApplication,
	report func(error),
) (*CustomerProvisioningHTTP, error) {
	if application == nil || report == nil {
		return nil, fmt.Errorf("customer provisioning HTTP dependencies are required")
	}
	return &CustomerProvisioningHTTP{application: application, report: report}, nil
}

func customerProvisioningOperationID(value string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(value)
	return parsed, err == nil && parsed != uuid.Nil && parsed.String() == value
}
