package devjournal

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

type clock interface{ Now() time.Time }

type Registry struct {
	path        string
	registryKID string
	privateKey  ed25519.PrivateKey
	resolver    ProposalKeyResolver
	clock       clock
	mutex       sync.Mutex
}

func New(
	path string,
	registryKID string,
	privateKey ed25519.PrivateKey,
	resolver ProposalKeyResolver,
	clock clock,
) (*Registry, error) {
	if !filepath.IsAbs(path) || registryKID == "" ||
		len(privateKey) != ed25519.PrivateKeySize || resolver == nil || clock == nil {
		return nil, fmt.Errorf("development authority journal configuration is invalid")
	}
	return &Registry{
		path: path, registryKID: registryKID,
		privateKey: append(ed25519.PrivateKey(nil), privateKey...),
		resolver:   resolver, clock: clock,
	}, nil
}

func (registry *Registry) withLock(ctx context.Context, callback func() error) error {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	lockPath := registry.path + ".lock"
	lock, err := acquireFileLock(lockPath)
	if err != nil {
		return err
	}
	defer releaseFileLock(lock, lockPath)
	return callback()
}
