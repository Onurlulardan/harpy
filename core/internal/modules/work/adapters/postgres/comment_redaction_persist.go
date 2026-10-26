package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/redact_comment.sql
var redactCommentSQL string

func (store *Store) PersistCommentRedaction(ctx context.Context, comment domain.Comment, expected int64) error {
	if !comment.Valid() || comment.RedactedAt == nil || expected < 1 || comment.RowVersion != expected+1 {
		return domain.ErrCommentRedaction
	}
	return executeOne(ctx, store, redactCommentSQL, comment.ID, comment.Redaction,
		comment.RedactedBy, comment.RedactionReason, comment.RowVersion, comment.RedactedAt, expected)
}
