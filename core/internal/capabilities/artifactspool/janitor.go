package artifactspool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

const OrphanGrace = time.Hour

type ActiveLeaseLookup interface {
	Active(context.Context, ReservationIdentity) (bool, error)
}

type ReapResult struct {
	Removed int
	Active  int
}

type reapCandidate struct {
	document     lockDocument
	lockPath     string
	spoolPath    string
	identityPath string
	active       bool
}

func (manager *Manager) Reap(
	ctx context.Context,
	lookup ActiveLeaseLookup,
	now time.Time,
) (ReapResult, error) {
	return manager.reap(ctx, lookup, now, OrphanGrace)
}

func (manager *Manager) reap(
	ctx context.Context,
	lookup ActiveLeaseLookup,
	now time.Time,
	grace time.Duration,
) (ReapResult, error) {
	if manager == nil || lookup == nil || now.IsZero() || grace < 0 {
		return ReapResult{}, fmt.Errorf("artifact spool janitor input is invalid")
	}
	candidates := make([]reapCandidate, 0, manager.slots)
	for slot := 0; slot < manager.slots; slot++ {
		document, found, err := manager.readLock(slot)
		if err != nil {
			return ReapResult{}, err
		}
		if !found || now.Sub(document.AcquiredAt) < grace {
			continue
		}
		active, err := lookup.Active(ctx, document.Identity)
		if err != nil {
			return ReapResult{}, fmt.Errorf("resolve artifact spool lease: %w", err)
		}
		candidates = append(candidates, reapCandidate{document: document,
			lockPath: manager.lockPath(slot), spoolPath: manager.spoolPath(slot),
			identityPath: manager.identityPath(document.Identity), active: active})
	}
	var result ReapResult
	for _, candidate := range candidates {
		if candidate.active {
			result.Active++
			continue
		}
		if err := removeCandidate(candidate); err != nil {
			return result, err
		}
		result.Removed++
	}
	return result, nil
}

func (manager *Manager) readLock(slot int) (lockDocument, bool, error) {
	file, err := os.Open(manager.lockPath(slot))
	if errors.Is(err, os.ErrNotExist) {
		return lockDocument{}, false, nil
	}
	if err != nil {
		return lockDocument{}, false, fmt.Errorf("open artifact spool lock: %w", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, 2049))
	decoder.DisallowUnknownFields()
	var document lockDocument
	if err := decoder.Decode(&document); err != nil || document.Version != 1 ||
		document.Slot != slot || document.Identity.validate() != nil ||
		document.AcquiredAt.IsZero() {
		return lockDocument{}, false, fmt.Errorf("artifact spool lock is invalid")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return lockDocument{}, false, fmt.Errorf("artifact spool lock has trailing data")
	}
	return document, true, nil
}

func removeCandidate(candidate reapCandidate) error {
	for _, path := range []string{
		candidate.spoolPath, candidate.lockPath, candidate.identityPath,
	} {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove orphan artifact spool: %w", err)
		}
	}
	return nil
}
