package artifactspool

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

var (
	ErrCapacity        = errors.New("artifact spool capacity is exhausted")
	ErrAlreadyReserved = errors.New("artifact spool attempt is already reserved")
)

type lockDocument struct {
	Version    int                 `json:"version"`
	Identity   ReservationIdentity `json:"identity"`
	Slot       int                 `json:"slot"`
	AcquiredAt time.Time           `json:"acquired_at"`
}

func (manager *Manager) Acquire(identity ReservationIdentity) (*Lease, error) {
	if manager == nil || identity.validate() != nil {
		return nil, fmt.Errorf("artifact spool reservation is invalid")
	}
	identityPath := manager.identityPath(identity)
	identityLock, err := os.OpenFile(identityPath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return nil, ErrAlreadyReserved
	}
	if err != nil {
		return nil, fmt.Errorf("create artifact spool attempt lock: %w", err)
	}
	if err := identityLock.Close(); err != nil {
		_ = os.Remove(identityPath)
		return nil, fmt.Errorf("close artifact spool attempt lock: %w", err)
	}
	keepIdentity := false
	defer func() {
		if !keepIdentity {
			_ = os.Remove(identityPath)
		}
	}()
	for slot := 0; slot < manager.slots; slot++ {
		lease, err := manager.acquireSlot(slot, identity, identityPath)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		keepIdentity = true
		return lease, nil
	}
	return nil, ErrCapacity
}

func (manager *Manager) acquireSlot(
	slot int,
	identity ReservationIdentity,
	identityPath string,
) (*Lease, error) {
	lockPath, spoolPath := manager.lockPath(slot), manager.spoolPath(slot)
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, err
	}
	keep := false
	defer func() {
		_ = lock.Close()
		if !keep {
			_ = os.Remove(spoolPath)
			_ = os.Remove(lockPath)
		}
	}()
	document := lockDocument{Version: 1, Identity: identity, Slot: slot,
		AcquiredAt: time.Now().UTC()}
	encoder := json.NewEncoder(lock)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(document); err != nil {
		return nil, fmt.Errorf("write artifact spool lock: %w", err)
	}
	if err := lock.Sync(); err != nil {
		return nil, fmt.Errorf("sync artifact spool lock: %w", err)
	}
	spool, err := os.OpenFile(spoolPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create artifact spool: %w", err)
	}
	if err := allocateFile(spool, manager.reservationBytes); err != nil {
		_ = spool.Close()
		return nil, fmt.Errorf("allocate artifact spool: %w", err)
	}
	keep = true
	return &Lease{manager: manager, identity: identity, slot: slot,
		lockPath: lockPath, identityPath: identityPath,
		spoolPath: spoolPath, file: spool}, nil
}
