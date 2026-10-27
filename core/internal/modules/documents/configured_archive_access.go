package documents

import (
	"context"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type DocumentsArchiveAccess interface {
	RefreshDocumentsArchive(context.Context, string, string, uint64, []string) error
}

type configuredArchiveAccess struct{ upstream DocumentsArchiveAccess }

func (access configuredArchiveAccess) Refresh(ctx context.Context,
	tenantID, actorID foundation.ID, revision uint64, itemIDs []foundation.ID,
) error {
	items := make([]string, len(itemIDs))
	for index, itemID := range itemIDs {
		items[index] = itemID.String()
	}
	if access.upstream == nil || access.upstream.RefreshDocumentsArchive(ctx,
		tenantID.String(), actorID.String(), revision, items) != nil {
		return archiveapp.ErrForbidden
	}
	return nil
}

var _ archiveapp.RefreshAuthorizer = configuredArchiveAccess{}
