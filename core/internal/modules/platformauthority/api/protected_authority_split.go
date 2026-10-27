package api

import "context"

type ProtectedAuthorityPlanner interface {
	PrepareChange(context.Context, ProtectedChangeRequest) (PreparedAuthority, error)
}
