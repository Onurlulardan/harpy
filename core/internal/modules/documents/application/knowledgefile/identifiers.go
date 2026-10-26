package knowledgefile

import "github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"

func (application *Application) nextIDs(count int) ([]domain.ID, error) {
	if application == nil || count < 1 || count > 3 {
		return nil, errUnavailable
	}
	result := make([]domain.ID, 0, count)
	seen := map[domain.ID]struct{}{}
	for len(result) < count {
		value, err := application.ids.NewUUIDv7()
		if _, duplicate := seen[value]; err != nil || !value.UUIDv7() || duplicate {
			return nil, errUnavailable
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}
