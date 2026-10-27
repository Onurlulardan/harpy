package knowledgefile

import (
	"context"
	"time"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	knowledgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/knowledgefile"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) VerifyOpen(ctx context.Context, itemID, versionID domain.ID,
	now time.Time,
) (knowledgeapp.Lease, error) {
	if !storage.usable(ctx) || !itemID.UUIDv7() || !versionID.UUIDv7() || now.IsZero() {
		return knowledgeapp.Lease{}, knowledgeapi.ErrInvalid
	}
	lease, err := scanLease(storage.queryRow(ctx, verifyOpenSQL,
		storage.actorID.String(), itemID.String(), versionID.String()), storage.tenantID)
	return storage.result(lease, err)
}
