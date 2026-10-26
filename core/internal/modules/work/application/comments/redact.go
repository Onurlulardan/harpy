package comments

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) RedactComment(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, commentID domain.ID, moderatorReason string,
) (ports.OperationReceipt, error) {
	moderatorReason = strings.TrimSpace(moderatorReason)
	if !commentID.Valid() || meta.ExpectedVersion < 1 || len(moderatorReason) > 1000 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(commentID.String(), moderatorReason, ports.VersionString(meta.ExpectedVersion))
	return service.mutator.Mutate(ctx, scope, "work.comment.redact", commentID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			comment, err := stores.Collaboration().LoadCommentForUpdate(ctx, commentID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if comment.RowVersion != meta.ExpectedVersion {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			moderator := comment.AuthorID != scope.ActorID()
			comment, err = comment.Redact(scope.ActorID(), moderator, moderatorReason, service.clock.Now().UTC())
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if err = stores.Collaboration().PersistCommentRedaction(ctx, comment, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				CommentID, Redaction string
				Version              int64
			}{
				commentID.String(), string(comment.Redaction), comment.RowVersion})
			return ports.OperationReceipt{AggregateID: commentID, RowVersion: comment.RowVersion}, payload, nil
		})
}
