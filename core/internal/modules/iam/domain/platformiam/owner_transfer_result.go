package platformiam

type OwnerTransferPrepared struct {
	OperationID, RegistryBindingID, ScopeID ID
	Generation                              int64
	PreviousHeadSHA256, StateSHA256         [32]byte
	ProposalSHA256                          [32]byte
	Status                                  string
	Replayed                                bool
}

func (result OwnerTransferPrepared) Validate(operation ID, expectedStatus string) error {
	if result.OperationID != operation || !result.RegistryBindingID.Valid() ||
		!result.ScopeID.Valid() || result.Generation < 1 || result.PreviousHeadSHA256 == [32]byte{} ||
		result.StateSHA256 == [32]byte{} || result.ProposalSHA256 == [32]byte{} ||
		result.Status != expectedStatus {
		return ErrStorage
	}
	return nil
}

type OwnerTransferCommitted struct {
	OperationID, PredecessorUserID, SuccessorUserID          ID
	PredecessorTokenVersion, PredecessorAuthorizationVersion int64
	SuccessorTokenVersion, SuccessorAuthorizationVersion     int64
	OwnerPolicyRevision, PlatformClock, HeadGeneration       int64
	Replayed                                                 bool
}

func (result OwnerTransferCommitted) Validate(operation ID) error {
	if result.OperationID != operation || !result.PredecessorUserID.Valid() ||
		!result.SuccessorUserID.Valid() || result.PredecessorUserID == result.SuccessorUserID ||
		result.PredecessorTokenVersion < 2 || result.PredecessorAuthorizationVersion < 1 ||
		result.SuccessorTokenVersion < 2 || result.SuccessorAuthorizationVersion < 1 ||
		result.OwnerPolicyRevision < 2 || result.PlatformClock < 2 || result.HeadGeneration < 2 {
		return ErrStorage
	}
	return nil
}

type OwnerTransferCompensated struct {
	OperationID, OriginalOperationID ID
	Generation                       int64
	StateSHA256, AckSHA256           [32]byte
	Replayed                         bool
}

func (result OwnerTransferCompensated) Validate(operation, original ID) error {
	if result.OperationID != operation || result.OriginalOperationID != original ||
		result.Generation < 3 || result.StateSHA256 == [32]byte{} || result.AckSHA256 == [32]byte{} {
		return ErrStorage
	}
	return nil
}
