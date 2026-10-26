package v1adoption

import (
	"context"
	"crypto/sha256"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

type objectEvidence struct{ source, target [32]byte }

func (store *TargetStore) ComparePlan(ctx context.Context, scope ports.Scope,
	plan adoption.Plan,
) (adoption.Evidence, error) {
	if err := store.check(ctx, scope); err != nil {
		return adoption.Evidence{}, err
	}
	if len(plan.Issues) > 0 {
		return adoption.Evidence{RunID: plan.RunID}, nil
	}
	objects, err := store.loadObjects(ctx, plan.RunID)
	if err != nil {
		return adoption.Evidence{}, err
	}
	if len(objects) != len(plan.Records) {
		return adoption.Evidence{}, adoption.ErrBlocked
	}
	for _, record := range plan.Records {
		value, exists := objects[objectKey(record.Kind, record.ID.String())]
		if !exists || value.source != record.SourceSHA256 || value.target != record.SHA256 {
			return adoption.Evidence{}, adoption.ErrBlocked
		}
	}
	targetHash := plan.TargetSHA256()
	acceptance := acceptanceDigest(plan.ManifestSHA256, plan.SourceSHA256, targetHash)
	if err = store.executeOne(ctx, acceptAdoptionSQL, plan.RunID, acceptance[:],
		len(plan.Records), targetHash[:]); err != nil {
		return adoption.Evidence{}, err
	}
	return adoption.Evidence{RunID: plan.RunID, SourceCount: plan.SourceCount,
		TargetCount: int64(len(plan.Records)), SourceSHA256: plan.SourceSHA256,
		TargetSHA256: targetHash, Accepted: true, WriterFenced: true}, nil
}

func (store *TargetStore) loadObjects(ctx context.Context, runID domain.ID) (
	map[string]objectEvidence, error,
) {
	rows, err := store.database.Query(ctx, listObjectsSQL, runID)
	if err != nil {
		return nil, ports.ErrUnavailable
	}
	defer rows.Close()
	result := map[string]objectEvidence{}
	for rows.Next() {
		var kind string
		var id uuid.UUID
		var sourceHash, targetHash []byte
		if rows.Scan(&kind, &id, &sourceHash, &targetHash) != nil ||
			len(sourceHash) != 32 || len(targetHash) != 32 {
			return nil, ports.ErrUnavailable
		}
		var value objectEvidence
		copy(value.source[:], sourceHash)
		copy(value.target[:], targetHash)
		result[objectKey(adoption.TargetKind(kind), id.String())] = value
	}
	if rows.Err() != nil {
		return nil, ports.ErrUnavailable
	}
	return result, nil
}

func acceptanceDigest(values ...[sha256.Size]byte) [sha256.Size]byte {
	hash := sha256.New()
	for _, value := range values {
		_, _ = hash.Write(value[:])
	}
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}
