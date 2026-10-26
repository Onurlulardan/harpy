package api

import "time"

type OnboardingPlanInput struct {
	OperationID             UUID
	EnvironmentID           UUID
	TenantID                UUID
	FirstUserID             UUID
	RootOrganizationID      UUID
	ProtectedAdminBindingID UUID
	InstanceOwnerBindingID  UUID
	Kind                    OnboardingKind
	DeploymentMode          DeploymentMode
	BindingPurpose          TenantBindingPurpose
	SchemaName              string
	IdempotencyKey          string
	RequestSHA256           string
	ManifestSHA256          string
	Offering                BaseOfferingPin
	RecoverySigning         RecoverySigningPin
	Authority               ProtectedAuthorityPlan
	CreatedAt               time.Time
}

type BaseOfferingPin struct {
	VersionID   UUID
	Revision    int64
	SHA256      string
	EffectiveAt time.Time
	Source      OfferingSourcePin
}

type OfferingSourcePin struct {
	Sequence int64
	SHA256   string
	Present  bool
}

type RecoverySigningPin struct {
	Generation        int64
	TrustBundleSHA256 string
	Present           bool
}

type ProtectedAuthorityPlan struct {
	RegistryBindingID  UUID
	ScopeKind          AuthorityScopeKind
	ScopeID            UUID
	Generation         int64
	PreviousHeadSHA256 string
	StateSHA256        string
	ProposalSHA256     string
}
