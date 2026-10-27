package recoveryanchorpostgres

import (
	"context"
	_ "embed"
	"time"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
)

//go:embed sql/platform/record_proposal.sql
var recordProposalSQL string

func (storage *store) PersistProposal(
	ctx context.Context,
	scope tenantScope,
	subject anchordomain.Subject,
	prepared anchordomain.PreparedAuthority,
	now time.Time,
) error {
	bound, err := storage.bind(ctx, scope, commandMode)
	if err != nil || scope.ActorID() != subject.ActorID || scope.TenantID() != subject.TenantID {
		return anchordomain.ErrInvalid
	}
	var recorded bool
	err = bound.database.QueryRow(ctx, recordProposalSQL,
		prepared.OperationID, subject.TenantID, subject.ActorID,
		prepared.RegistryBindingID, prepared.Generation, prepared.StateSHA256[:],
		prepared.ProposalSHA256[:], now).Scan(&recorded)
	if err != nil {
		return storeError(err, storage.rowNotFound)
	}
	if !recorded {
		return anchordomain.ErrConflict
	}
	return nil
}
