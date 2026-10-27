package api

import (
	"time"
)

func (plan OnboardingPlan) OperationID() UUID        { return plan.input.OperationID }
func (plan OnboardingPlan) EnvironmentID() UUID      { return plan.input.EnvironmentID }
func (plan OnboardingPlan) TenantID() UUID           { return plan.input.TenantID }
func (plan OnboardingPlan) FirstUserID() UUID        { return plan.input.FirstUserID }
func (plan OnboardingPlan) RootOrganizationID() UUID { return plan.input.RootOrganizationID }
func (plan OnboardingPlan) ProtectedAdminBindingID() UUID {
	return plan.input.ProtectedAdminBindingID
}
func (plan OnboardingPlan) InstanceOwnerBindingID() UUID {
	return plan.input.InstanceOwnerBindingID
}
func (plan OnboardingPlan) Kind() OnboardingKind                 { return plan.input.Kind }
func (plan OnboardingPlan) DeploymentMode() DeploymentMode       { return plan.input.DeploymentMode }
func (plan OnboardingPlan) BindingPurpose() TenantBindingPurpose { return plan.input.BindingPurpose }
func (plan OnboardingPlan) SchemaName() string                   { return plan.input.SchemaName }
func (plan OnboardingPlan) IdempotencyKey() string               { return plan.input.IdempotencyKey }
func (plan OnboardingPlan) RequestSHA256() string                { return plan.input.RequestSHA256 }
func (plan OnboardingPlan) ManifestSHA256() string               { return plan.input.ManifestSHA256 }
func (plan OnboardingPlan) StepPlanSHA256() string               { return plan.stepPlanSHA256 }
func (plan OnboardingPlan) Offering() BaseOfferingPin            { return plan.input.Offering }
func (plan OnboardingPlan) RecoverySigning() RecoverySigningPin  { return plan.input.RecoverySigning }
func (plan OnboardingPlan) Authority() ProtectedAuthorityPlan    { return plan.input.Authority }
func (plan OnboardingPlan) CreatedAt() time.Time                 { return plan.input.CreatedAt }

func (plan OnboardingPlan) IsValid() bool {
	rebuilt, err := NewOnboardingPlan(plan.input)
	return err == nil && rebuilt.stepPlanSHA256 == plan.stepPlanSHA256
}
