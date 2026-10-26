package notificationhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
)

var errRealtimeUnavailable = errors.New("notification realtime transport is unavailable")

type socketKey struct{ tenant, user string }

type Hub struct {
	mutex           sync.RWMutex
	connections     map[socketKey]map[*socketConnection]struct{}
	connectionCount int
	accepting       bool
	empty           chan struct{}
}

func NewHub() *Hub {
	empty := make(chan struct{})
	close(empty)
	return &Hub{connections: make(map[socketKey]map[*socketConnection]struct{}),
		accepting: true, empty: empty}
}

func (hub *Hub) Serve(w http.ResponseWriter, r *http.Request,
	identity application.RealtimeConnectionIdentity,
) error {
	if hub == nil || w == nil || r == nil || !identity.TenantID.Valid() || !identity.UserID.Valid() {
		return errRealtimeUnavailable
	}
	connection, err := upgradeSocket(w, r)
	if err != nil {
		return errRealtimeUnavailable
	}
	key := socketKey{identity.TenantID.String(), identity.UserID.String()}
	if !hub.add(key, connection) {
		connection.close()
		return errRealtimeUnavailable
	}
	defer func() {
		hub.remove(key, connection)
		connection.close()
	}()
	return connection.run()
}

func (hub *Hub) Broadcast(ctx context.Context, hint RealtimeHint) error {
	if hub == nil || ctx == nil || ctx.Err() != nil || !hint.TenantID.Valid() ||
		!hint.UserID.Valid() || !hint.MessageID.Valid() || hint.InboxRevision < 1 {
		return errRealtimeUnavailable
	}
	body, err := json.Marshal(struct {
		Type          string `json:"type"`
		MessageID     string `json:"message_id"`
		InboxRevision int64  `json:"inbox_revision"`
	}{"notification_available", hint.MessageID.String(), hint.InboxRevision})
	if err != nil {
		return fmt.Errorf("encode notification hint: %w", errRealtimeUnavailable)
	}
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()
	for connection := range hub.connections[socketKey{hint.TenantID.String(), hint.UserID.String()}] {
		connection.offer(body)
	}
	return nil
}

func (hub *Hub) add(key socketKey, connection *socketConnection) bool {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	if !hub.accepting {
		return false
	}
	if _, exists := hub.connections[key][connection]; exists {
		return false
	}
	if hub.connectionCount == 0 {
		hub.empty = make(chan struct{})
	}
	if hub.connections[key] == nil {
		hub.connections[key] = make(map[*socketConnection]struct{})
	}
	hub.connections[key][connection] = struct{}{}
	hub.connectionCount++
	return true
}

func (hub *Hub) remove(key socketKey, connection *socketConnection) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	if _, found := hub.connections[key][connection]; !found {
		return
	}
	delete(hub.connections[key], connection)
	hub.connectionCount--
	if len(hub.connections[key]) == 0 {
		delete(hub.connections, key)
	}
	if hub.connectionCount == 0 {
		close(hub.empty)
	}
}

var _ RealtimeTransport = (*Hub)(nil)
