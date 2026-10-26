package archive

import (
	"context"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) FindCreate(ctx context.Context, key string,
	digest [32]byte,
) (archiveapp.Snapshot, bool, error) {
	if !storage.usable(ctx) || storage.job || key == "" || digest == [32]byte{} {
		return archiveapp.Snapshot{}, false, archiveapp.ErrInvalid
	}
	var id string
	err := storage.queryRow(ctx, findCreateSQL, storage.actorID.String(), key,
		digest[:]).Scan(&id)
	if err != nil {
		if storage.notFound(err) {
			return archiveapp.Snapshot{}, false, nil
		}
		return archiveapp.Snapshot{}, false, archiveapp.ErrUnavailable
	}
	parsed, err := domain.ParseID(id)
	if err != nil {
		return archiveapp.Snapshot{}, false, archiveapp.ErrUnavailable
	}
	result, err := storage.Get(ctx, parsed)
	return result, err == nil, err
}
