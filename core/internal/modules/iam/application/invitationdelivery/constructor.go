package invitationdelivery

import "reflect"

type Service struct{ store Store }

func NewService(store Store) (*Service, error) {
	if nilStore(store) {
		return nil, ErrInvalid
	}
	return &Service{store: store}, nil
}

func nilStore(store Store) bool {
	if store == nil {
		return true
	}
	value := reflect.ValueOf(store)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
