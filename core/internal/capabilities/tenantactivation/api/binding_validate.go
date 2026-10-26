package api

import "github.com/google/uuid"

func (binding Binding) valid() bool {
	identities := []uuid.UUID{binding.OperationID, binding.EnvironmentID, binding.TenantID,
		binding.FirstUserID, binding.RootOrganizationID, binding.ProtectedAdminBindingID,
		binding.Offering.VersionID, binding.Authority.RegistryBindingID,
		binding.Authority.ScopeID, binding.MigrationReceipt.ID}
	for _, identity := range identities {
		if identity == uuid.Nil {
			return false
		}
	}
	if !validProfile(binding) || binding.SchemaName == "" ||
		!digestPattern.MatchString(binding.RequestSHA256) ||
		!digestPattern.MatchString(binding.ManifestSHA256) ||
		!digestPattern.MatchString(binding.StepPlanSHA256) ||
		!digestPattern.MatchString(binding.MigrationReceipt.SHA256) ||
		!validOffering(binding.Offering) || !validRecovery(binding) ||
		!validAuthority(binding) {
		return false
	}
	ownerRequired := binding.Kind == "environment_initialize"
	return ownerRequired && binding.InstanceOwnerBindingID != uuid.Nil ||
		!ownerRequired && binding.InstanceOwnerBindingID == uuid.Nil
}

func validProfile(binding Binding) bool {
	if binding.Kind == "saas_customer_tenant_provision" {
		return binding.DeploymentMode == "saas" && binding.BindingPurpose == "saas_customer"
	}
	if binding.Kind != "environment_initialize" {
		return false
	}
	return binding.DeploymentMode == "onpremise" &&
		binding.BindingPurpose == "onpremise_primary" ||
		binding.DeploymentMode == "saas" && binding.BindingPurpose == "saas_control"
}

func validOffering(pin OfferingPin) bool {
	return pin.Revision > 0 && digestPattern.MatchString(pin.SHA256) &&
		!pin.EffectiveAt.IsZero() &&
		(!pin.Source.Present && pin.Source.Sequence == 0 && pin.Source.SHA256 == "" ||
			pin.Source.Present && pin.Source.Sequence > 0 &&
				digestPattern.MatchString(pin.Source.SHA256))
}

func validRecovery(binding Binding) bool {
	pin := binding.RecoverySigning
	if binding.DeploymentMode == "onpremise" {
		return pin.Present && pin.Generation > 0 &&
			digestPattern.MatchString(pin.TrustBundleSHA256) && binding.Offering.Source.Present
	}
	return !pin.Present && pin.Generation == 0 && pin.TrustBundleSHA256 == ""
}

func validAuthority(binding Binding) bool {
	pin := binding.Authority
	if binding.Kind == "saas_customer_tenant_provision" {
		return validAuthorityGenerationOne(pin, "tenant", binding.TenantID)
	}
	if binding.DeploymentMode == "saas" {
		return validAuthorityGenerationOne(pin, "environment", binding.EnvironmentID)
	}
	return pin.ScopeKind == "environment" && pin.ScopeID == binding.EnvironmentID &&
		pin.Generation == 0 && pin.PreviousHeadSHA256 == "" && pin.StateSHA256 == "" &&
		pin.ProposalSHA256 == "" && pin.AckSHA256 == ""
}

func validAuthorityGenerationOne(pin AuthorityPin, kind string, scopeID uuid.UUID) bool {
	return pin.ScopeKind == kind && pin.ScopeID == scopeID && pin.Generation == 1 &&
		pin.PreviousHeadSHA256 == "" && digestPattern.MatchString(pin.StateSHA256) &&
		digestPattern.MatchString(pin.ProposalSHA256) && digestPattern.MatchString(pin.AckSHA256)
}
