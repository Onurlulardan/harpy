package application

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func receiptFrom(operation ports.OperationRecord) api.AuthorityReceipt {
	return api.AuthorityReceipt{
		OperationID:       operation.OperationID,
		RegistryBindingID: operation.Proposal.Value.RegistryBindingID,
		ScopeKind:         string(operation.Proposal.Value.ScopeKind),
		ScopeID:           operation.Proposal.Value.ScopeID,
		Generation:        operation.Proposal.Value.Generation,
		StateSHA256:       operation.State.SHA256,
		ProposalSHA256:    operation.Proposal.SHA256,
		AckSHA256:         operation.Ack.SHA256,
		RegistrySequence:  operation.Ack.Value.RegistrySequence,
		CompletedAt:       operation.CompletedAt,
	}
}

func mapStoreError(err error) error {
	if errors.Is(err, ports.ErrStaleHead) {
		return api.ErrStaleHead
	}
	if errors.Is(err, ports.ErrConflict) {
		return api.ErrConflict
	}
	return api.ErrDependency
}
