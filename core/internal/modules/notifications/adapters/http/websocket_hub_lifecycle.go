package notificationhttp

import (
	"context"
	"sort"
)

func (hub *Hub) ActiveTenantIDs() []string {
	if hub == nil {
		return nil
	}
	hub.mutex.RLock()
	tenants := make(map[string]struct{}, len(hub.connections))
	for key := range hub.connections {
		tenants[key.tenant] = struct{}{}
	}
	hub.mutex.RUnlock()
	values := make([]string, 0, len(tenants))
	for tenant := range tenants {
		values = append(values, tenant)
	}
	sort.Strings(values)
	return values
}

func (hub *Hub) BeginDrain() {
	for _, connection := range hub.closeAdmissionAndSnapshot() {
		connection.close()
	}
}

func (hub *Hub) Wait(ctx context.Context) error {
	if hub == nil || ctx == nil {
		return errRealtimeUnavailable
	}
	hub.mutex.RLock()
	empty := hub.empty
	hub.mutex.RUnlock()
	select {
	case <-empty:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (hub *Hub) Force() error {
	if hub == nil {
		return errRealtimeUnavailable
	}
	hub.BeginDrain()
	return nil
}

func (hub *Hub) closeAdmissionAndSnapshot() []*socketConnection {
	if hub == nil {
		return nil
	}
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	hub.accepting = false
	result := make([]*socketConnection, 0, hub.connectionCount)
	for _, connections := range hub.connections {
		for connection := range connections {
			result = append(result, connection)
		}
	}
	return result
}
