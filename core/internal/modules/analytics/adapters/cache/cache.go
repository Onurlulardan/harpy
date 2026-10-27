package cache

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
)

type Clock interface{ Now() time.Time }

type Cache struct {
	mutex   sync.Mutex
	clock   Clock
	limit   int
	entries map[[32]byte]entry
	queries map[uuid.UUID]map[[32]byte]struct{}
}

type entry struct {
	queryID   uuid.UUID
	body      []byte
	expiresAt time.Time
}

func New(clock Clock, limit int) (*Cache, error) {
	if nilPort(clock) || limit < 1 || limit > 65536 {
		return nil, fmt.Errorf("analytics cache configuration is invalid")
	}
	return &Cache{clock: clock, limit: limit, entries: make(map[[32]byte]entry),
		queries: make(map[uuid.UUID]map[[32]byte]struct{})}, nil
}

func nilPort(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}

var _ queryapp.Cache = (*Cache)(nil)
