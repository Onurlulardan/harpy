package identity

type Change struct {
	OperationID   IdentityID
	ActorID       IdentityID
	AggregateID   IdentityID
	RequestSHA256 [32]byte
	ResourceKind  string
	CommandKey    string
	ResultCode    string
}

func NewChange(
	evidence MutationEvidence,
	aggregateID IdentityID,
	resourceKind string,
	commandKey string,
) Change {
	return Change{
		OperationID: evidence.OperationID, ActorID: evidence.ActorID,
		AggregateID: aggregateID, RequestSHA256: evidence.RequestSHA256,
		ResourceKind: resourceKind, CommandKey: commandKey,
		ResultCode: commandKey + ".succeeded",
	}
}

func (change Change) Validate() error {
	if change.OperationID == (IdentityID{}) || change.ActorID == (IdentityID{}) ||
		change.AggregateID == (IdentityID{}) || change.RequestSHA256 == [32]byte{} ||
		!identityChangeResource(change.ResourceKind) ||
		!identityChangeCommand(change.ResourceKind, change.CommandKey) ||
		change.ResultCode != change.CommandKey+".succeeded" {
		return ErrChangeSink
	}
	return nil
}

func identityChangeCommand(resourceKind string, commandKey string) bool {
	switch resourceKind {
	case "user":
		return commandKey == "iam.identity.user.create" ||
			commandKey == "iam.identity.user.update" ||
			commandKey == "iam.identity.user.disable" ||
			commandKey == "iam.identity.user.delete"
	case "role":
		return commandKey == "iam.identity.role.create" ||
			commandKey == "iam.identity.role.update" ||
			commandKey == "iam.identity.role.retire"
	case "organization":
		return commandKey == "iam.identity.organization.create" ||
			commandKey == "iam.identity.organization.update" ||
			commandKey == "iam.identity.organization.disable" ||
			commandKey == "iam.identity.organization.delete"
	default:
		return false
	}
}

func identityChangeResource(value string) bool {
	return value == "user" || value == "role" || value == "organization"
}

type RecordedChange struct {
	Change
	Revision      int64
	Clock         int64
	EventSequence int64
}

func (record RecordedChange) Validate() error {
	if record.Change.Validate() != nil || record.Revision < 1 ||
		record.Clock < 1 || record.EventSequence < 1 {
		return ErrChangeSink
	}
	return nil
}
