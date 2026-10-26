package upload

import "github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"

func (application *Application) nextIDs(count int) ([]domain.ID, error) {
	if count < 1 || count > 8 {
		return nil, ErrUnavailable
	}
	result := make([]domain.ID, 0, count)
	seen := make(map[domain.ID]struct{}, count)
	for range count {
		value, err := application.ids.NewUUIDv7()
		if err != nil || !value.UUIDv7() {
			return nil, ErrUnavailable
		}
		if _, exists := seen[value]; exists {
			return nil, ErrUnavailable
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}
