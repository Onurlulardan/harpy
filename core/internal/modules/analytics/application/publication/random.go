package publication

import "github.com/google/uuid"

func (application *Application) nextID() (uuid.UUID, error) {
	value, err := application.ids.NewUUIDv7()
	if err != nil || !validUUIDv7(value) {
		return uuid.Nil, ErrUnavailable
	}
	return value, nil
}

func (application *Application) randomBytes(size int) ([]byte, error) {
	value := make([]byte, size)
	if err := application.random.Read(value); err != nil {
		return nil, ErrUnavailable
	}
	zero := true
	for _, item := range value {
		zero = zero && item == 0
	}
	if zero {
		return nil, ErrUnavailable
	}
	return value, nil
}
