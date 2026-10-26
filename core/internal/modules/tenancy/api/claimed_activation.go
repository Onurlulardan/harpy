package api

import "time"

type ClaimedActivation struct {
	OperationID            UUID
	TenantID               UUID
	Kind                   OnboardingKind
	StepPlanSHA256         string
	MigrationReceiptID     UUID
	MigrationReceiptSHA256 string
	AuthorityAckSHA256     string
	Revision               int64
	DatabaseNow            time.Time
	LeaseUntil             time.Time
}

func (claim ClaimedActivation) IsValid() bool {
	kindValid := claim.Kind == OnboardingEnvironmentInitialize ||
		claim.Kind == OnboardingSaaSCustomerProvision
	return claim.OperationID != NilUUID && claim.TenantID != NilUUID && kindValid &&
		validOnboardingDigest(claim.StepPlanSHA256) && claim.MigrationReceiptID != NilUUID &&
		validOnboardingDigest(claim.MigrationReceiptSHA256) && claim.Revision > 0 &&
		microsecondUTC(claim.DatabaseNow) && microsecondUTC(claim.LeaseUntil) &&
		claim.LeaseUntil.After(claim.DatabaseNow) &&
		(claim.AuthorityAckSHA256 == "" || validOnboardingDigest(claim.AuthorityAckSHA256))
}
