package api

import "context"

type InitialAuthorityPlanner interface {
	PrepareInitial(context.Context, InitialAuthorityRequest) (PreparedAuthority, error)
}

type InitialAuthorityAcknowledger interface {
	AcknowledgePrepared(context.Context, string) (AuthorityReceipt, error)
}
