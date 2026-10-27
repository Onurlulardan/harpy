package v1adoption

import (
	"context"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (store *TargetStore) SealLegacy(ctx context.Context, scope ports.Scope,
	request adoption.RetireRequest, evidence adoption.Evidence,
) (adoption.Receipt, error) {
	if err := store.check(ctx, scope); err != nil {
		return adoption.Receipt{}, err
	}
	if !request.Valid() || request.RunID != evidence.RunID || !evidence.Accepted ||
		!evidence.WriterFenced || evidence.CompatibilityHits != 0 {
		return adoption.Receipt{}, ports.ErrInvalid
	}
	if err := store.executeOne(ctx, retireWriterSQL, request.RunID, scope.ActorID()); err != nil {
		return adoption.Receipt{}, err
	}
	return adoption.Receipt{RunID: request.RunID, State: "RETIRED",
		SourceCount: evidence.SourceCount, TargetCount: evidence.TargetCount,
		SHA256: evidence.TargetSHA256}, nil
}
