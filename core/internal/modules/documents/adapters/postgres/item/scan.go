package item

import (
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func scanSnapshot(value row, tenantID domain.ID,
	notFound rowNotFound,
) (itemapp.Snapshot, error) {
	var document []byte
	if err := value.Scan(&document); err != nil {
		if notFound(err) {
			return itemapp.Snapshot{}, itemapp.ErrNotFound
		}
		return itemapp.Snapshot{}, itemapp.ErrUnavailable
	}
	return decodeItem(tenantID, document)
}

func scanMutation(value row, tenantID domain.ID,
	notFound rowNotFound,
) (itemapp.Snapshot, bool, error) {
	var document []byte
	var replay bool
	if err := value.Scan(&document, &replay); err != nil {
		if notFound(err) {
			return itemapp.Snapshot{}, false, itemapp.ErrConflict
		}
		return itemapp.Snapshot{}, false, itemapp.ErrUnavailable
	}
	result, err := decodeItem(tenantID, document)
	return result, replay, err
}
