package comments

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) CreateComment(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, boardID, taskID domain.ID, body string,
) (ports.OperationReceipt, error) {
	body = strings.TrimSpace(body)
	if !boardID.Valid() || !taskID.Valid() || body == "" || len(body) > 16_000 || meta.ExpectedVersion != 0 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(boardID.String(), taskID.String(), body)
	return service.mutator.Mutate(ctx, scope, "work.comment.create", taskID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			task, err := stores.Tasks().LoadTaskForUpdate(ctx, taskID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if task.BoardID != boardID || task.DeletedAt != nil {
				return ports.OperationReceipt{}, nil, ports.ErrNotFound
			}
			now := service.clock.Now().UTC()
			id, err := service.ids.New(now)
			if err != nil {
				return ports.OperationReceipt{}, nil, ports.ErrUnavailable
			}
			comment := domain.Comment{ID: id, BoardID: boardID, TaskID: taskID,
				AuthorID: scope.ActorID(), Body: body, CreatedAt: now, RowVersion: 1}
			if !comment.Valid() {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			if err = stores.Collaboration().InsertComment(ctx, comment); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct{ BoardID, TaskID, CommentID string }{
				boardID.String(), taskID.String(), id.String()})
			return ports.OperationReceipt{AggregateID: id, RowVersion: 1}, payload, nil
		})
}
