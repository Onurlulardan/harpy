package transaction

type ScopeRequest struct {
	Actor     VerifiedActor
	Operation VerifiedOperation
	Mode      ScopeMode
}

func (request ScopeRequest) valid() bool {
	return request.Actor.valid() && request.Operation.valid() && request.Mode.valid() &&
		request.Operation.mode == request.Mode
}
