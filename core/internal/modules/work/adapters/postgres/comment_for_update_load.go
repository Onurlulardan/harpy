package postgres

import (
	"context"
	_ "embed"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/get_comment_for_update.sql
var getCommentForUpdateSQL string

func (store *Store) LoadCommentForUpdate(ctx context.Context, id domain.ID) (domain.Comment, error) {
	if !id.Valid() {
		return domain.Comment{}, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return domain.Comment{}, err
	}
	var commentID, boardID, taskID, authorID uuid.UUID
	var body, state string
	var kind, reason *string
	var redactedBy uuid.NullUUID
	var version int64
	var createdAt time.Time
	var redactedAt *time.Time
	err := store.database.QueryRow(ctx, getCommentForUpdateSQL, id).Scan(&commentID, &boardID, &taskID,
		&authorID, &body, &state, &kind, &redactedBy, &reason, &version, &createdAt, &redactedAt)
	if err != nil {
		return domain.Comment{}, err
	}
	comment := domain.Comment{ID: domain.ID(commentID), BoardID: domain.ID(boardID), TaskID: domain.ID(taskID),
		AuthorID: domain.ID(authorID), Body: body, RowVersion: version, CreatedAt: createdAt.UTC()}
	if redactedAt != nil {
		if kind == nil {
			return domain.Comment{}, application.ErrUnavailable
		}
		value := redactedAt.UTC()
		comment.RedactedAt = &value
		comment.RedactedBy = domain.ID(redactedBy.UUID)
		comment.Redaction = domain.RedactionKind(*kind)
		if reason != nil {
			comment.RedactionReason = *reason
		}
	}
	if !comment.Valid() || state != map[bool]string{true: "REDACTED", false: "VISIBLE"}[comment.RedactedAt != nil] {
		return domain.Comment{}, application.ErrUnavailable
	}
	return comment, nil
}
