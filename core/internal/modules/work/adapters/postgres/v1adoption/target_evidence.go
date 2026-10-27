package v1adoption

import (
	"context"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

func (store *TargetStore) LoadEvidence(ctx context.Context, scope ports.Scope,
	runID domain.ID,
) (adoption.Evidence, error) {
	if err := store.check(ctx, scope); err != nil {
		return adoption.Evidence{}, err
	}
	if !runID.Valid() {
		return adoption.Evidence{}, ports.ErrInvalid
	}
	var id uuid.UUID
	var sourceCount, targetCount, writes, hits int64
	var sourceHash, targetHash []byte
	var accepted, fenced bool
	err := store.database.QueryRow(ctx, readEvidenceSQL, runID).Scan(&id,
		&sourceCount, &targetCount, &sourceHash, &targetHash, &accepted, &fenced, &writes, &hits)
	if err != nil || domain.ID(id) != runID || len(sourceHash) != 32 || len(targetHash) != 32 ||
		sourceCount < 0 || targetCount < 0 || writes < 0 || hits < 0 {
		return adoption.Evidence{}, ports.ErrUnavailable
	}
	result := adoption.Evidence{RunID: runID, SourceCount: sourceCount,
		TargetCount: targetCount, Accepted: accepted, WriterFenced: fenced,
		V2WriteCount: writes, CompatibilityHits: hits}
	copy(result.SourceSHA256[:], sourceHash)
	copy(result.TargetSHA256[:], targetHash)
	return result, nil
}
