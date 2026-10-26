package recoveryset

func ValidateRestore(request RestoreRequest) error {
	manifest := request.Manifest
	if !validUUID(request.RecoverySetID) || !manifest.valid() ||
		manifest.Portability != SameEnvironment ||
		request.SeedRequested || request.OwnerChange || request.TenantReplacement ||
		!sameIdentity(manifest.Identity, request.Target) {
		return ErrRestoreGate
	}
	switch request.Disposition {
	case TargetExistingReadable:
		if !validUUID(request.Safety.RecoverySetID) ||
			request.Safety.RecoverySetID == request.RecoverySetID ||
			!digestPattern.MatchString(request.Safety.RecoverySetSHA256) ||
			request.Safety.WaiverReason != "" || request.Safety.ProbeSHA256 != "" ||
			request.Safety.DestructiveConfirmed {
			return ErrRestoreGate
		}
	case TargetEmpty:
		if !validWaiver(request.Safety, "empty_no_database_or_policy", false) {
			return ErrRestoreGate
		}
	case TargetUnreadable:
		if !validWaiver(request.Safety, "unreadable_integrity_probe", true) {
			return ErrRestoreGate
		}
	default:
		return ErrRestoreGate
	}
	return nil
}

func sameIdentity(source, target EnvironmentIdentity) bool {
	return source.EnvironmentID == target.EnvironmentID &&
		source.DeploymentMode == ModeOnPremise && target.DeploymentMode == ModeOnPremise &&
		source.PrimaryBindingID == target.PrimaryBindingID &&
		source.TenantSetSHA256 == target.TenantSetSHA256 &&
		source.OwnerBindingID == target.OwnerBindingID &&
		source.OwnerBindingGeneration == target.OwnerBindingGeneration &&
		source.DatabaseLineageID == target.DatabaseLineageID
}

func validWaiver(evidence SafetyEvidence, reason string, destructive bool) bool {
	return evidence.RecoverySetID == "" && evidence.RecoverySetSHA256 == "" &&
		evidence.WaiverReason == reason && digestPattern.MatchString(evidence.ProbeSHA256) &&
		evidence.DestructiveConfirmed == destructive
}
