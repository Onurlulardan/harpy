package artifactspool

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Lease struct {
	manager      *Manager
	identity     ReservationIdentity
	slot         int
	lockPath     string
	identityPath string
	spoolPath    string
	file         *os.File
	mu           sync.Mutex
	opened       bool
	sealed       bool
	discarded    bool
}

func (lease *Lease) OpenWriter() (io.WriteCloser, error) {
	lease.mu.Lock()
	defer lease.mu.Unlock()
	if lease.discarded || lease.opened || lease.sealed || lease.file == nil {
		return nil, fmt.Errorf("artifact spool writer state is invalid")
	}
	if _, err := lease.file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek artifact spool: %w", err)
	}
	lease.opened = true
	return &boundedWriter{lease: lease}, nil
}

func (lease *Lease) OpenReader() (io.ReadCloser, error) {
	lease.mu.Lock()
	defer lease.mu.Unlock()
	if lease.discarded || !lease.sealed || lease.file != nil {
		return nil, fmt.Errorf("artifact spool reader state is invalid")
	}
	file, err := os.Open(lease.spoolPath)
	if err != nil {
		return nil, fmt.Errorf("open artifact spool reader: %w", err)
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 ||
		info.Size() > lease.manager.reservationBytes {
		_ = file.Close()
		return nil, fmt.Errorf("artifact spool file is invalid")
	}
	return file, nil
}

func (lease *Lease) Discard() error {
	lease.mu.Lock()
	defer lease.mu.Unlock()
	if lease.discarded {
		return nil
	}
	if lease.file != nil {
		_ = lease.file.Close()
		lease.file = nil
	}
	spoolErr := os.Remove(lease.spoolPath)
	if os.IsNotExist(spoolErr) {
		spoolErr = nil
	}
	lockErr := os.Remove(lease.lockPath)
	if os.IsNotExist(lockErr) {
		lockErr = nil
	}
	identityErr := os.Remove(lease.identityPath)
	if os.IsNotExist(identityErr) {
		identityErr = nil
	}
	lease.discarded = spoolErr == nil && lockErr == nil && identityErr == nil
	return errorsJoin(errorsJoin(spoolErr, lockErr), identityErr)
}

func errorsJoin(first, second error) error {
	if first != nil {
		return first
	}
	return second
}
