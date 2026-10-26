package tenancy

import (
	"time"

	tenancyapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type DeploymentMode = tenancyapi.DeploymentMode
type PrimaryTenantPurpose = tenancyapi.PrimaryTenantPurpose

const (
	DeploymentModeOnPremise       = tenancyapi.DeploymentModeOnPremise
	DeploymentModeSaaS            = tenancyapi.DeploymentModeSaaS
	PrimaryTenantPurposeOnPremise = tenancyapi.PrimaryTenantPurposeOnPremise
	PrimaryTenantPurposeSaaS      = tenancyapi.PrimaryTenantPurposeSaaS
)

func NewDeploymentPolicy(
	environmentID string,
	mode DeploymentMode,
	maximumCustomerTenants *int64,
	createdAt time.Time,
) (DeploymentPolicy, error) {
	return tenancyapi.NewDeploymentPolicy(
		environmentID,
		mode,
		maximumCustomerTenants,
		createdAt,
	)
}

func NewPrimaryTenantTarget(
	environmentID string,
	mode DeploymentMode,
	tenantID string,
	purpose PrimaryTenantPurpose,
	publicSlug string,
) (PrimaryTenantTarget, error) {
	return tenancyapi.NewPrimaryTenantTarget(
		environmentID,
		mode,
		tenantID,
		purpose,
		publicSlug,
	)
}
