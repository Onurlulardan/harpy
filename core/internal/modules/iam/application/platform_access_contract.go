package application

import (
	platformiam "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/sessioncapability"
)

type PlatformDecision = platformiam.Decision
type PlatformEvaluateQuery = platformiam.EvaluateQuery
type PlatformSessionContextQuery = platformiam.SessionContextQuery
type PlatformID = platformiam.ID

type DeploymentMode = sessioncapability.DeploymentMode
type TenantPurpose = sessioncapability.TenantPurpose
type SessionCapabilitySets = sessioncapability.Sets

const (
	DeploymentModeOnPremise = sessioncapability.ModeOnPremise
	DeploymentModeSaaS      = sessioncapability.ModeSaaS

	TenantPurposeOnPremisePrimary = sessioncapability.PurposeOnPremisePrimary
	TenantPurposeSaaSControl      = sessioncapability.PurposeSaaSControl
	TenantPurposeSaaSCustomer     = sessioncapability.PurposeSaaSCustomer
)

var PlatformCapabilityCatalog = platformiam.CapabilityCatalog

func ModeAllowsPlatform(mode DeploymentMode, capability string) bool {
	return sessioncapability.ModeAllowsPlatform(mode, capability)
}
