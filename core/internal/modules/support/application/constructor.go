package application

import (
	"fmt"
	"reflect"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	supportports "github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

type Service struct {
	store Store
	clock Clock
	ids   IDGenerator
}

func NewService(store Store, clock Clock, ids IDGenerator) (*Service, error) {
	if nilPort(store) || nilPort(clock) || nilPort(ids) {
		return nil, fmt.Errorf("support application dependencies are required")
	}
	return &Service{store: store, clock: clock, ids: ids}, nil
}

func (service *Service) now() (time.Time, error) {
	if service == nil || nilPort(service.clock) {
		return time.Time{}, ErrUnavailable
	}
	now := service.clock.Now().UTC().Truncate(time.Microsecond)
	if now.IsZero() || now.Nanosecond()%1000 != 0 {
		return time.Time{}, ErrUnavailable
	}
	return now, nil
}

func (service *Service) newID() (domain.ID, error) {
	if service == nil || nilPort(service.ids) {
		return domain.NilID, ErrUnavailable
	}
	id, err := service.ids.NewV7()
	if err != nil || id == domain.NilID || id.Version() != 7 {
		return domain.NilID, ErrUnavailable
	}
	return id, nil
}

func nilPort(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func validCommandEvidence(key string, digest [32]byte) bool {
	return supportports.ValidCommandEvidence(key, digest)
}
