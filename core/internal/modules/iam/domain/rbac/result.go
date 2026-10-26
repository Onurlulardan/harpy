package rbac

type MutationResult struct {
	ID                 ID
	Revision           int64
	AuthorizationClock int64
	Replayed           bool
}

func (result MutationResult) Validate(expected ID) error {
	if !result.ID.Valid() || result.ID != expected || result.Revision < 1 ||
		result.AuthorizationClock < 1 {
		return ErrStorage
	}
	return nil
}

type Change struct {
	OperationID   ID
	ActorID       ID
	AggregateID   ID
	RequestSHA256 [32]byte
	CommandKey    string
	ResultCode    string
	Revision      int64
	Clock         int64
}

func (change Change) Validate() error {
	if !change.OperationID.Valid() || !change.ActorID.Valid() ||
		!change.AggregateID.Valid() || change.RequestSHA256 == [32]byte{} ||
		change.CommandKey == "" || change.ResultCode == "" ||
		change.Revision < 1 || change.Clock < 1 {
		return ErrSink
	}
	return nil
}

func NewChange(
	result MutationResult,
	evidence MutationEvidence,
	commandKey string,
	resultCode string,
) Change {
	return Change{
		OperationID: evidence.OperationID, ActorID: evidence.ActorID,
		AggregateID: result.ID, RequestSHA256: evidence.RequestSHA256,
		CommandKey: commandKey, ResultCode: resultCode,
		Revision: result.Revision, Clock: result.AuthorizationClock,
	}
}
