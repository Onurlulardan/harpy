package projections

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) ConsumeProjectionEvent(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, item ports.ProjectionItem,
) (ports.OperationReceipt, error) {
	if !validProjectionInput(item, true) {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := projectionFingerprint(item, true)
	return service.mutator.Mutate(ctx, scope, "work.projection.apply", item.SourceID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			cursor, err := stores.Projections().LockCursor(ctx, item.SourceOwner, item.SourceKind, item.SourceID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			_, disposition, err := cursor.Accept(item.EventSequence)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if disposition == domain.ProjectionDuplicate {
				return projectionReceipt(item, disposition)
			}
			if disposition == domain.ProjectionGap {
				if err = stores.Projections().FlagRepair(ctx, item.SourceOwner, item.SourceKind,
					item.SourceID, item.EventSequence); err != nil {
					return ports.OperationReceipt{}, nil, err
				}
				return projectionReceipt(item, disposition)
			}
			allowed, currentRevision, err := service.resources.AuthorizeRead(ctx, scope,
				ports.ResourceAuthorityRequest{Owner: item.SourceOwner, Kind: item.SourceKind,
					ID: item.SourceID, VersionID: item.SourceVersionID, Epoch: item.ReferenceEpoch})
			if err != nil || currentRevision < 1 {
				return ports.OperationReceipt{}, nil, ports.ErrUnavailable
			}
			decision := domain.DecideProjectionVisibility(domain.ProjectionItem{ID: item.ID,
				SourceID: item.SourceID, SourceVersionID: item.SourceVersionID, Owner: item.SourceOwner,
				SourceRevision: item.SourceRevision, PermissionRevision: item.PermissionRevision,
				EventSequence: item.EventSequence, Summary: item.Summary}, allowed, currentRevision)
			item.Visible = decision.Visible
			if err = stores.Projections().PersistProjection(ctx, item); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if decision.RepairNeeded {
				if err = stores.Projections().FlagRepair(ctx, item.SourceOwner, item.SourceKind,
					item.SourceID, item.EventSequence); err != nil {
					return ports.OperationReceipt{}, nil, err
				}
				return projectionReceipt(item, domain.ProjectionGap)
			}
			if err = stores.Projections().AdvanceCursor(ctx, item.SourceOwner, item.SourceKind,
				item.SourceID, item.EventSequence); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			return projectionReceipt(item, disposition)
		})
}

func validProjectionPrincipals(principals []domain.Principal) bool {
	if len(principals) == 0 || len(principals) > 100 {
		return false
	}
	seen := make(map[domain.Principal]struct{}, len(principals))
	for _, principal := range principals {
		if !principal.Valid() {
			return false
		}
		if _, duplicate := seen[principal]; duplicate {
			return false
		}
		seen[principal] = struct{}{}
	}
	return true
}

func projectionReceipt(item ports.ProjectionItem, disposition domain.ProjectionDisposition) (ports.OperationReceipt, []byte, error) {
	payload, _ := json.Marshal(struct {
		SourceID, Disposition string
		Sequence              int64
	}{
		item.SourceID.String(), string(disposition), item.EventSequence})
	return ports.OperationReceipt{AggregateID: item.SourceID, RowVersion: item.EventSequence}, payload, nil
}
