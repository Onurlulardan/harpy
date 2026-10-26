package api

import (
	"context"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

type customerProvisioningSetup struct {
	preparer *iammodule.InitialAdminSetupPreparer
	reissuer *iammodule.InitialAdminSetupReissuer
}

func (setup customerProvisioningSetup) PrepareInitialAdmin(
	ctx context.Context,
	request tenancymodule.CustomerProvisioningSetupRequest,
) error {
	if setup.preparer == nil {
		return fmt.Errorf("initial admin setup preparer unavailable")
	}
	identifierID := iammodule.InstallationFirstEmailIdentifierID(request.OperationID)
	if identifierID == [16]byte{} {
		return fmt.Errorf("initial admin setup identifier is invalid")
	}
	_, err := setup.preparer.Prepare(ctx, iammodule.InitialAdminSetupPrepareInput{
		OperationID: request.OperationID, TenantID: request.TenantID,
		UserID: request.UserID, IdentifierID: identifierID,
		Email: request.Email, CreatedAt: request.CreatedAt,
	})
	return err
}

func (setup customerProvisioningSetup) ReissueInitialAdmin(
	ctx context.Context,
	request tenancymodule.CustomerProvisioningSetupReissueRequest,
) error {
	if setup.reissuer == nil {
		return fmt.Errorf("initial admin setup reissuer unavailable")
	}
	return setup.reissuer.Reissue(ctx, iammodule.InitialAdminSetupReissueInput{
		OperationID:        request.OperationID,
		ReissueOperationID: request.ReissueOperationID,
		TenantID:           request.TenantID, UserID: request.UserID,
		SchemaName: request.SchemaName, Email: request.Email,
		CreatedAt: request.CreatedAt,
	})
}

var _ tenancymodule.CustomerProvisioningSetupPreparer = customerProvisioningSetup{}
