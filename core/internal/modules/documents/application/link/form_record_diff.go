package link

import (
	"slices"

	formrecordlink "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordlink"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type formRecordMutation struct {
	itemID  domain.ID
	release *domain.Link
}

func canonicalFormRecordRequest(
	request formrecordlink.Request,
) (Source, []domain.ID, error) {
	source := Source{Owner: formrecordlink.SourceOwner, Kind: formrecordlink.SourceKind,
		ID: domain.ID(request.DefinitionID), VersionID: domain.ID(request.RecordID)}
	seen := make(map[domain.ID]struct{}, len(request.DesiredItemIDs))
	desired := make([]domain.ID, 0, len(request.DesiredItemIDs))
	for _, raw := range request.DesiredItemIDs {
		itemID := domain.ID(raw)
		if !itemID.UUIDv7() {
			return Source{}, nil, formrecordlink.ErrInvalid
		}
		if _, duplicate := seen[itemID]; !duplicate {
			seen[itemID] = struct{}{}
			desired = append(desired, itemID)
		}
	}
	slices.SortFunc(desired, compareDomainID)
	return source, desired, nil
}

func formRecordDiff(
	tenantID domain.ID,
	source Source,
	current []domain.Link,
	desired []domain.ID,
) ([]formRecordMutation, error) {
	active := make(map[domain.ID]domain.Link, len(current))
	for _, link := range current {
		if link.TenantID != tenantID || link.SourceOwner != source.Owner ||
			link.SourceKind != source.Kind || link.SourceID != source.ID ||
			link.SourceVersionID != source.VersionID ||
			link.Usage != formrecordlink.Usage || link.State != domain.LinkActive ||
			!link.ItemID.UUIDv7() {
			return nil, formrecordlink.ErrUnavailable
		}
		if _, duplicate := active[link.ItemID]; duplicate {
			return nil, formrecordlink.ErrUnavailable
		}
		active[link.ItemID] = link
	}
	wanted := make(map[domain.ID]struct{}, len(desired))
	for _, itemID := range desired {
		wanted[itemID] = struct{}{}
	}
	ids := make([]domain.ID, 0, len(active)+len(wanted))
	for itemID := range active {
		ids = append(ids, itemID)
	}
	for itemID := range wanted {
		if _, exists := active[itemID]; !exists {
			ids = append(ids, itemID)
		}
	}
	slices.SortFunc(ids, compareDomainID)
	result := make([]formRecordMutation, 0, len(ids))
	for _, itemID := range ids {
		link, exists := active[itemID]
		_, keep := wanted[itemID]
		switch {
		case exists && !keep:
			copy := link
			result = append(result, formRecordMutation{itemID: itemID, release: &copy})
		case !exists && keep:
			result = append(result, formRecordMutation{itemID: itemID})
		}
	}
	return result, nil
}

func compareDomainID(left, right domain.ID) int {
	return slices.Compare(left[:], right[:])
}
