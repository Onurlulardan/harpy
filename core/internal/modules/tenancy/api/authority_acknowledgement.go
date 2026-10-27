package api

import "time"

type AuthorityAcknowledgement struct {
	OperationID        UUID
	RegistryBindingID  UUID
	ScopeKind          AuthorityScopeKind
	ScopeID            UUID
	Generation         int64
	PreviousHeadSHA256 string
	StateSHA256        string
	ProposalSHA256     string
	AckSHA256          string
	RegistrySequence   int64
	VerifiedAt         time.Time
}

func (ack AuthorityAcknowledgement) IsValid() bool {
	return ack.OperationID != NilUUID && ack.RegistryBindingID != NilUUID &&
		(ack.ScopeKind == AuthorityScopeEnvironment || ack.ScopeKind == AuthorityScopeTenant) &&
		ack.ScopeID != NilUUID && ack.Generation > 0 && ack.PreviousHeadSHA256 == "" &&
		validOnboardingDigest(ack.StateSHA256) && validOnboardingDigest(ack.ProposalSHA256) &&
		validOnboardingDigest(ack.AckSHA256) && ack.RegistrySequence > 0 &&
		microsecondUTC(ack.VerifiedAt)
}
