package api

import (
	"time"

	"github.com/google/uuid"
)

type OfferingSourcePin struct {
	Sequence int64
	SHA256   string
	Present  bool
}

type OfferingPin struct {
	VersionID   uuid.UUID
	Revision    int64
	SHA256      string
	EffectiveAt time.Time
	Source      OfferingSourcePin
}

type RecoverySigningPin struct {
	Generation        int64
	TrustBundleSHA256 string
	Present           bool
}

type AuthorityPin struct {
	RegistryBindingID  uuid.UUID
	ScopeKind          string
	ScopeID            uuid.UUID
	Generation         int64
	PreviousHeadSHA256 string
	StateSHA256        string
	ProposalSHA256     string
	AckSHA256          string
}

type MigrationReceiptPin struct {
	ID     uuid.UUID
	SHA256 string
}

type Binding struct {
	OperationID             uuid.UUID
	EnvironmentID           uuid.UUID
	TenantID                uuid.UUID
	FirstUserID             uuid.UUID
	RootOrganizationID      uuid.UUID
	ProtectedAdminBindingID uuid.UUID
	InstanceOwnerBindingID  uuid.UUID
	Kind                    string
	DeploymentMode          string
	BindingPurpose          string
	SchemaName              string
	RequestSHA256           string
	ManifestSHA256          string
	StepPlanSHA256          string
	Offering                OfferingPin
	RecoverySigning         RecoverySigningPin
	Authority               AuthorityPin
	MigrationReceipt        MigrationReceiptPin
}
