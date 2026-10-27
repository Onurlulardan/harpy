package notificationhttp

import (
	"bufio"
	"context"
	"net"
	"reflect"
	"testing"
)

func TestHubTracksActiveTenantsAndDrainsConnections(t *testing.T) {
	hub := NewHub()
	firstKey := socketKey{
		tenant: "019d0000-0000-7000-8000-000000000002",
		user:   "019d0000-0000-7000-8000-000000000102",
	}
	secondKey := socketKey{
		tenant: "019d0000-0000-7000-8000-000000000001",
		user:   "019d0000-0000-7000-8000-000000000101",
	}
	first := lifecycleSocket(t)
	second := lifecycleSocket(t)
	if !hub.add(firstKey, first) || !hub.add(secondKey, second) {
		t.Fatal("active connections were not admitted")
	}
	want := []string{secondKey.tenant, firstKey.tenant}
	if got := hub.ActiveTenantIDs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("active tenants = %v, want %v", got, want)
	}
	hub.BeginDrain()
	for _, connection := range []*socketConnection{first, second} {
		select {
		case <-connection.done:
		default:
			t.Fatal("connection remained open during drain")
		}
	}
	if hub.add(firstKey, lifecycleSocket(t)) {
		t.Fatal("connection admitted after drain")
	}
	hub.remove(firstKey, first)
	hub.remove(secondKey, second)
	if err := hub.Wait(context.Background()); err != nil {
		t.Fatalf("wait drained hub: %v", err)
	}
}

func lifecycleSocket(t *testing.T) *socketConnection {
	t.Helper()
	server, client := net.Pipe()
	t.Cleanup(func() { _ = client.Close() })
	return newSocketConnection(server, bufio.NewReader(server))
}
