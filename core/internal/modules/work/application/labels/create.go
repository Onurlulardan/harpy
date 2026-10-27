package labels

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) CreateLabel(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, boardID domain.ID, name, color string,
) (ports.OperationReceipt, error) {
	name, color = strings.TrimSpace(name), strings.ToLower(strings.TrimSpace(color))
	if !boardID.Valid() || name == "" || len(name) > 100 || meta.ExpectedVersion != 0 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	return service.mutator.Mutate(ctx, scope, "work.label.change", boardID, meta,
		ports.Fingerprint(boardID.String(), name, color),
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			id, err := service.ids.New(service.clock.Now().UTC())
			if err != nil {
				return ports.OperationReceipt{}, nil, ports.ErrUnavailable
			}
			label := domain.Label{ID: id, BoardID: boardID, Name: name, Color: color, RowVersion: 1}
			if !label.Valid() {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			if err = stores.Tasks().InsertLabel(ctx, label); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct{ BoardID, LabelID string }{boardID.String(), id.String()})
			return ports.OperationReceipt{AggregateID: id, RowVersion: 1}, payload, nil
		})
}
