package api

import (
	"context"
	"time"
)

type ChangePurpose string

const (
	ChangeOwnerTransfer  ChangePurpose = "owner_transfer"
	ChangeProtectedAdmin ChangePurpose = "protected_admin_change"
	ChangeRecoveryAnchor ChangePurpose = "recovery_anchor_change"
)

type ProtectedChangeRequest struct {
	OperationID string
	Purpose     ChangePurpose
	State       State
	IssuedAt    time.Time
	Policy      ExpiryPolicy
}

type ProtectedAuthority interface {
	PrepareChange(context.Context, ProtectedChangeRequest) (PreparedAuthority, error)
	AcknowledgeChange(context.Context, string) (AuthorityReceipt, error)
}
