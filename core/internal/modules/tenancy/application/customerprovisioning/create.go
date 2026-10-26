package customerprovisioning

import (
	"context"

	provisioningv1 "github.com/Onurlulardan/harpy/contracts/customerprovisioning/v1"
)

const CommandProvisionCustomerTenant = provisioningv1.CommandProvisionCustomerTenant

type Command struct {
	Caller         CallerKind
	IdempotencyKey string
	Customer       CustomerRequest
	Credential     CredentialEnvelope
}

func (coordinator *Coordinator) Create(
	ctx context.Context,
	command Command,
) (Operation, error) {
	return coordinator.provision(ctx, command)
}
