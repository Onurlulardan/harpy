package link

import (
	formrecordlink "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordlink"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (registry *FormRecordRegistry) nextIDs(count int) ([]domain.ID, error) {
	if registry == nil || count < 1 || count > 3 {
		return nil, formrecordlink.ErrUnavailable
	}
	result := make([]domain.ID, 0, count)
	seen := map[domain.ID]struct{}{}
	for len(result) < count {
		value, err := registry.ids.NewUUIDv7()
		if _, duplicate := seen[value]; err != nil || !value.UUIDv7() || duplicate {
			return nil, formrecordlink.ErrUnavailable
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}
