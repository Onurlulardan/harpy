package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/insert_comment.sql
var insertCommentSQL string

func (store *Store) InsertComment(ctx context.Context, comment domain.Comment) error {
	if !comment.Valid() || comment.RedactedAt != nil {
		return domain.ErrCommentRedaction
	}
	return executeOne(ctx, store, insertCommentSQL, comment.ID, comment.BoardID, comment.TaskID,
		comment.AuthorID, comment.Body, comment.RowVersion, comment.CreatedAt)
}
