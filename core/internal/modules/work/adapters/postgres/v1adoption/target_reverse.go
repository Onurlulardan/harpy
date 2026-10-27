package v1adoption

import (
	"context"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (store *TargetStore) RevertPlan(ctx context.Context, scope ports.Scope,
	request adoption.ReverseRequest, evidence adoption.Evidence,
) (adoption.Receipt, error) {
	if err := store.check(ctx, scope); err != nil {
		return adoption.Receipt{}, err
	}
	if !request.Valid() || request.RunID != evidence.RunID ||
		request.ApprovedBy == scope.ActorID() || !evidence.Accepted || !evidence.WriterFenced {
		return adoption.Receipt{}, ports.ErrInvalid
	}
	if err := store.writer.Revert(ctx, request, evidence); err != nil {
		return adoption.Receipt{}, err
	}
	if err := store.executeOne(ctx, acceptReverseSQL, request.ReconciliationID,
		request.SourceGeneration, request.TargetGeneration, evidence.SourceCount,
		evidence.TargetCount, evidence.SourceSHA256[:], evidence.TargetSHA256[:],
		scope.ActorID(), request.ApprovedBy); err != nil {
		return adoption.Receipt{}, err
	}
	if err := store.executeOne(ctx, reopenWriterSQL, request.RunID,
		request.ReconciliationID, scope.ActorID()); err != nil {
		return adoption.Receipt{}, err
	}
	if err := store.executeOne(ctx, setAdoptionStateSQL, request.RunID, "REVERSED",
		[]string{"ACCEPTED", "REVERSED"}); err != nil {
		return adoption.Receipt{}, err
	}
	return adoption.Receipt{RunID: request.RunID, State: "REVERSED",
		SourceCount: evidence.SourceCount, TargetCount: evidence.TargetCount,
		SHA256: evidence.TargetSHA256}, nil
}
