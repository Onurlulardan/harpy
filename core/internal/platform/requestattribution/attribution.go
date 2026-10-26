package requestattribution

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrInvalid     = errors.New("request attribution identity is invalid")
	ErrUnavailable = errors.New("request attribution slot is unavailable")
	ErrConflict    = errors.New("request attribution identity conflicts")
)

type Status uint8

const (
	StatusUnavailable Status = iota
	StatusUnattributed
	StatusVerified
	StatusConflict
)

type Identity struct {
	ActorID, TenantID [16]byte
}

func (identity Identity) Valid() bool {
	return identity.ActorID != [16]byte{} && identity.TenantID != [16]byte{}
}

type contextKey struct{}

type slot struct {
	mu       sync.RWMutex
	identity Identity
	status   Status
}

func WithSlot(ctx context.Context) (context.Context, error) {
	if ctx == nil {
		return nil, ErrUnavailable
	}
	if existing, ok := ctx.Value(contextKey{}).(*slot); ok && existing != nil {
		return ctx, nil
	}
	return context.WithValue(ctx, contextKey{}, &slot{status: StatusUnattributed}), nil
}

func Bind(ctx context.Context, actorID, tenantID [16]byte) error {
	identity := Identity{ActorID: actorID, TenantID: tenantID}
	value := slotFromContext(ctx)
	if value == nil {
		return ErrUnavailable
	}
	if !identity.Valid() {
		return ErrInvalid
	}
	value.mu.Lock()
	defer value.mu.Unlock()
	switch value.status {
	case StatusUnattributed:
		value.identity, value.status = identity, StatusVerified
		return nil
	case StatusVerified:
		if value.identity == identity {
			return nil
		}
		value.identity, value.status = Identity{}, StatusConflict
		return ErrConflict
	case StatusConflict:
		return ErrConflict
	default:
		return ErrUnavailable
	}
}

func Snapshot(ctx context.Context) (Identity, Status) {
	value := slotFromContext(ctx)
	if value == nil {
		return Identity{}, StatusUnavailable
	}
	value.mu.RLock()
	defer value.mu.RUnlock()
	if value.status != StatusVerified {
		return Identity{}, value.status
	}
	return value.identity, value.status
}

func slotFromContext(ctx context.Context) *slot {
	if ctx == nil {
		return nil
	}
	value, _ := ctx.Value(contextKey{}).(*slot)
	return value
}
