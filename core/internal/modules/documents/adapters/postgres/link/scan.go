package link

import (
	"strings"

	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) scan(value row) (domain.Link, bool, error) {
	var result domain.Link
	var id, item, sourceID, sourceVersion, state string
	var released dbInstant
	var replay bool
	err := value.Scan(&id, &item, &result.SourceOwner, &result.SourceKind,
		&sourceID, &sourceVersion, &result.Usage, &result.ReferenceEpoch,
		&state, &result.CreatedAt, &released, &replay)
	if err != nil {
		if storage.notFound(err) {
			return domain.Link{}, false, linkapp.ErrConflict
		}
		return domain.Link{}, false, linkapp.ErrUnavailable
	}
	if result.ID, err = domain.ParseID(id); err == nil {
		result.ItemID, err = domain.ParseID(item)
	}
	if err == nil {
		result.SourceID, err = domain.ParseID(sourceID)
	}
	if err == nil {
		result.SourceVersionID, err = domain.ParseID(sourceVersion)
	}
	if err != nil {
		return domain.Link{}, false, linkapp.ErrUnavailable
	}
	result.TenantID, result.ReleasedAt = storage.tenantID, released.Time
	switch strings.ToUpper(state) {
	case "ACTIVE":
		result.State = domain.LinkActive
	case "RELEASED":
		result.State = domain.LinkReleased
	default:
		return domain.Link{}, false, linkapp.ErrUnavailable
	}
	return result, replay, nil
}
