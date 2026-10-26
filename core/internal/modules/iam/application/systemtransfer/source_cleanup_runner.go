package systemtransfer

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type SourceCleanupAdmission interface {
	Accepting() bool
}

type SourceCleanupRunner struct {
	store     SourceCleanupStore
	objects   SourceObjectDeleter
	clock     SourceCleanupClock
	jitter    SourceCleanupJitter
	ownerID   uuid.UUID
	admission SourceCleanupAdmission

	mutex       sync.Mutex
	started     bool
	scanCancel  context.CancelFunc
	forceCancel context.CancelFunc
	done        chan struct{}
	failures    chan error
}
