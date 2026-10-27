package labels

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) UpdateLabel(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, label domain.Label,
) (ports.OperationReceipt, error) {
	label.Name, label.Color = strings.TrimSpace(label.Name), strings.ToLower(strings.TrimSpace(label.Color))
	if !label.ID.Valid() || !label.BoardID.Valid() || meta.ExpectedVersion < 1 || label.DeletedAt != nil {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	label.RowVersion = meta.ExpectedVersion + 1
	if !label.Valid() {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(label.ID.String(), label.Name, label.Color,
		ports.VersionString(meta.ExpectedVersion))
	return service.mutator.Mutate(ctx, scope, "work.label.change", label.BoardID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			if err := stores.Tasks().PersistLabel(ctx, label, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				LabelID string
				Version int64
			}{label.ID.String(), label.RowVersion})
			return ports.OperationReceipt{AggregateID: label.ID, RowVersion: label.RowVersion}, payload, nil
		})
}
