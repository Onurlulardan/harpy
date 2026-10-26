package dashboard

import "github.com/google/uuid"

func (application *Application) nextIDs(count int) ([]uuid.UUID, error) {
	values := make([]uuid.UUID, count)
	seen := map[uuid.UUID]struct{}{}
	for index := range values {
		value, err := application.ids.NewUUIDv7()
		if err != nil || !validUUIDv7(value) {
			return nil, ErrUnavailable
		}
		if _, exists := seen[value]; exists {
			return nil, ErrUnavailable
		}
		seen[value] = struct{}{}
		values[index] = value
	}
	return values, nil
}
