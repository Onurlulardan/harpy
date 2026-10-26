package recoveryanchor

import "time"

type ID = [16]byte

type Subject struct {
	EnvironmentID ID
	TenantID      ID
	ActorID       ID
}

func (value Subject) Valid() bool {
	return value.EnvironmentID != (ID{}) && value.TenantID != (ID{}) &&
		value.ActorID != (ID{})
}

type Anchor struct {
	ID               ID
	Generation       int64
	CredentialID     []byte
	RPID             string
	UserVerification string
	BackupEligible   bool
	BackupState      bool
	State            string
}

type RegistrationOptions struct {
	OperationID ID
	AnchorID    ID
	Challenge   string
	RPID        string
	Origins     []string
	ExpiresAt   time.Time
	Replayed    bool
}

type OptionsCommand struct {
	Subject
	OperationID, AnchorID ID
	IdempotencyKey        string
	RequestSHA256         [32]byte
	Challenge, RPID       string
	Origins               []string
	OriginsSHA256         [32]byte
	ExpiresAt, OccurredAt time.Time
}

type RegistrationLookup struct {
	OperationID, AnchorID ID
	Challenge, RPID       string
	Origins               []string
	OriginsSHA256         [32]byte
	ExpiresAt             time.Time
	State                 string
}

type RegistrationEvidence struct {
	CredentialID, PublicKeyCOSE        []byte
	AAGUID                             ID
	AttestationFormat, AttestationType string
	SignCount                          int64
}

type RegisterCommand struct {
	Subject
	OperationID    ID
	IdempotencyKey string
	RequestSHA256  [32]byte
	Evidence       RegistrationEvidence
	RPID           string
	Origins        []string
	OriginsSHA256  [32]byte
	OccurredAt     time.Time
}

type DeleteCommand struct {
	Subject
	OperationID, AnchorID ID
	IdempotencyKey        string
	RequestSHA256         [32]byte
	OccurredAt            time.Time
}

type StagedMutation struct {
	OperationID, AnchorID ID
	AnchorSetSHA256       [32]byte
	Authority             AuthorityState
	Replayed              bool
}

type AuthorityState struct {
	EnvironmentID, OwnerBindingID, RegistryBindingID      ID
	DeploymentMode, ProposalSignerKID                     string
	OwnerBindingGeneration, ProposalSignerGeneration      int64
	PrimaryAdminBindingSetSHA256, RecoveryAnchorSetSHA256 [32]byte
	RecoveryTrustBundleSHA256                             string
}

type PreparedAuthority struct {
	OperationID       ID
	RegistryBindingID ID
	Generation        int64
	StateSHA256       [32]byte
	ProposalSHA256    [32]byte
}

type ProposedMutation struct {
	OperationID ID
	AnchorID    ID
	Kind        string
	State       string
}

type CommitResult struct {
	OperationID, AnchorID ID
	Kind                  string
	AuthorityGeneration   int64
	Replayed              bool
}
