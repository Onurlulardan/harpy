package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

type CollaborationStore interface {
	InsertComment(context.Context, domain.Comment) error
	LoadCommentForUpdate(context.Context, domain.ID) (domain.Comment, error)
	PersistCommentRedaction(context.Context, domain.Comment, int64) error
}
