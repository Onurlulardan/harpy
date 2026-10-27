package legalhold

import (
	"strings"

	holdapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/legalhold"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) scan(value row) (domain.LegalHold, bool, error) {
	var result domain.LegalHold
	var id, item, state string
	var evidence []byte
	var released dbInstant
	var replay bool
	err := value.Scan(&id, &item, &result.SourceOwner, &result.SourceCaseID,
		&evidence, &state, &result.PlacedAt, &released, &replay)
	if err != nil {
		if storage.notFound(err) {
			return domain.LegalHold{}, false, holdapp.ErrConflict
		}
		return domain.LegalHold{}, false, holdapp.ErrUnavailable
	}
	if result.ID, err = domain.ParseID(id); err == nil {
		result.ItemID, err = domain.ParseID(item)
	}
	if err != nil || len(evidence) != len(result.EvidenceSHA256) {
		return domain.LegalHold{}, false, holdapp.ErrUnavailable
	}
	copy(result.EvidenceSHA256[:], evidence)
	result.TenantID, result.ReleasedAt = storage.tenantID, released.Time
	switch strings.ToUpper(state) {
	case "ACTIVE":
		result.State = domain.LegalHoldActive
	case "RELEASED":
		result.State = domain.LegalHoldReleased
	default:
		return domain.LegalHold{}, false, holdapp.ErrUnavailable
	}
	return result, replay, nil
}
