package recoveryanchor

import (
	"context"
	"time"
)

type AuthorityState struct {
	EnvironmentID, OwnerBindingID, RegistryBindingID      string
	DeploymentMode, ProposalSignerKID                     string
	OwnerBindingGeneration, ProposalSignerGeneration      int64
	PrimaryAdminBindingSetSHA256, RecoveryAnchorSetSHA256 string
	RecoveryTrustBundleSHA256                             string
}

type PreparedAuthority struct {
	OperationID, RegistryBindingID string
	Generation                     int64
	StateSHA256, ProposalSHA256    string
}

type AuthorityPlanner interface {
	Prepare(context.Context, string, AuthorityState, time.Time) (PreparedAuthority, error)
}
