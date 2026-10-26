package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type connection struct {
	cancel context.CancelFunc
	close  func() error
	once   sync.Once
	err    error
}

func (entry *connection) stop() error {
	entry.once.Do(func() {
		entry.cancel()
		if entry.close != nil {
			entry.err = entry.close()
		}
	})
	return entry.err
}

type Connections struct {
	mu        sync.Mutex
	entries   map[*connection]struct{}
	accepting bool
	empty     chan struct{}
}

func NewConnections() *Connections {
	empty := make(chan struct{})
	close(empty)
	return &Connections{entries: make(map[*connection]struct{}), accepting: true, empty: empty}
}

func (connections *Connections) Track(
	parent context.Context,
	closeConnection func() error,
) (context.Context, func(), error) {
	if parent == nil {
		return nil, nil, fmt.Errorf("connection context is required")
	}
	child, cancel := context.WithCancel(parent)
	entry := &connection{cancel: cancel, close: closeConnection}
	connections.mu.Lock()
	defer connections.mu.Unlock()
	if !connections.accepting {
		cancel()
		return nil, nil, fmt.Errorf("long-lived connection admission is closed")
	}
	if len(connections.entries) == 0 {
		connections.empty = make(chan struct{})
	}
	connections.entries[entry] = struct{}{}
	var once sync.Once
	release := func() {
		once.Do(func() {
			entry.cancel()
			connections.remove(entry)
		})
	}
	return child, release, nil
}

func (connections *Connections) BeginDrain() error {
	entries := connections.snapshotAndCloseAdmission()
	var joined error
	for _, entry := range entries {
		joined = errors.Join(joined, entry.stop())
	}
	return joined
}

func (connections *Connections) Wait(ctx context.Context) error {
	connections.mu.Lock()
	empty := connections.empty
	connections.mu.Unlock()
	select {
	case <-empty:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (connections *Connections) Force() error {
	return connections.BeginDrain()
}

func (connections *Connections) Hook(name string) Hook {
	return Hook{Name: name, Begin: connections.BeginDrain, Wait: connections.Wait, Force: connections.Force}
}

func (connections *Connections) snapshotAndCloseAdmission() []*connection {
	connections.mu.Lock()
	defer connections.mu.Unlock()
	connections.accepting = false
	result := make([]*connection, 0, len(connections.entries))
	for entry := range connections.entries {
		result = append(result, entry)
	}
	return result
}

func (connections *Connections) remove(entry *connection) {
	connections.mu.Lock()
	defer connections.mu.Unlock()
	delete(connections.entries, entry)
	if len(connections.entries) == 0 {
		select {
		case <-connections.empty:
		default:
			close(connections.empty)
		}
	}
}
