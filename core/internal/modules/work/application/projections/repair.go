package projections

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) RepairProjection(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, owner domain.SourceOwner,
	kind string, sourceID domain.ID, through int64, items []ports.ProjectionItem,
) (ports.OperationReceipt, error) {
	if owner == domain.SourceWork || !owner.Valid() || kind == "" || !sourceID.Valid() ||
		through < 1 || len(items) > 10_000 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	for _, item := range items {
		if !validProjectionInput(item, false) || item.SourceOwner != owner ||
			item.SourceKind != kind || item.SourceID != sourceID || item.EventSequence > through {
			return ports.OperationReceipt{}, ports.ErrInvalid
		}
	}
	hash := repairFingerprint(owner, kind, sourceID, through, items)
	return service.mutator.Mutate(ctx, scope, "work.projection.repair", sourceID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			resolved := make([]ports.ProjectionItem, 0, len(items))
			for _, item := range items {
				allowed, revision, err := service.resources.AuthorizeRead(ctx, scope,
					ports.ResourceAuthorityRequest{Owner: owner, Kind: kind, ID: sourceID,
						VersionID: item.SourceVersionID, Epoch: item.ReferenceEpoch})
				if err != nil || revision < 1 {
					return ports.OperationReceipt{}, nil, ports.ErrUnavailable
				}
				item.PermissionRevision, item.Visible = revision, allowed
				resolved = append(resolved, item)
			}
			if err := stores.Projections().ReplaceSourceProjection(ctx, owner, kind, sourceID,
				resolved, through); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				SourceID string
				Through  int64
				Count    int
			}{
				sourceID.String(), through, len(resolved)})
			return ports.OperationReceipt{AggregateID: sourceID, RowVersion: through}, payload, nil
		})
}
