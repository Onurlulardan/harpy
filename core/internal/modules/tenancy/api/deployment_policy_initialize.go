package api

import "time"

type InitializeDeploymentPolicyRequest struct {
	OperationID    UUID
	EnvironmentID  UUID
	OperationKind  OnboardingKind
	DeploymentMode DeploymentMode
}

func (request InitializeDeploymentPolicyRequest) IsValid() bool {
	return request.OperationID != NilUUID && request.OperationID[6]>>4 == 7 &&
		request.EnvironmentID != NilUUID &&
		request.OperationKind == OnboardingEnvironmentInitialize &&
		(request.DeploymentMode == DeploymentModeOnPremise ||
			request.DeploymentMode == DeploymentModeSaaS)
}

type DeploymentPolicyInitialization struct {
	OperationID UUID
	Policy      DeploymentPolicy
	DatabaseNow time.Time
	Replayed    bool
}

func (receipt DeploymentPolicyInitialization) IsValid() bool {
	return receipt.OperationID != NilUUID && receipt.OperationID[6]>>4 == 7 &&
		receipt.Policy.IsValid() && microsecondUTC(receipt.DatabaseNow) &&
		!receipt.DatabaseNow.Before(receipt.Policy.CreatedAt())
}
