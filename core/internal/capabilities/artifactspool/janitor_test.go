package artifactspool

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type janitorLookup struct {
	active map[string]bool
	err    error
}

func (lookup janitorLookup) Active(
	_ context.Context,
	identity ReservationIdentity,
) (bool, error) {
	return lookup.active[identity.ExportID.String()], lookup.err
}

func TestJanitorRemovesOnlyInactiveLeaseAfterGrace(t *testing.T) {
	manager, _ := newManager(filepath.Join(t.TempDir(), "spool"), 64, 256)
	active, _ := manager.Acquire(spoolIdentity(1))
	orphan, _ := manager.Acquire(spoolIdentity(2))
	sealLease(t, active)
	sealLease(t, orphan)
	lookup := janitorLookup{active: map[string]bool{
		active.identity.ExportID.String(): true,
	}}
	result, err := manager.reap(context.Background(), lookup,
		time.Now().UTC().Add(time.Second), 0)
	if err != nil || result.Active != 1 || result.Removed != 1 {
		t.Fatalf("janitor result=%#v err=%v", result, err)
	}
	if _, err := os.Stat(active.spoolPath); err != nil {
		t.Fatalf("active spool removed: %v", err)
	}
	if _, err := os.Stat(orphan.spoolPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("orphan spool remains: %v", err)
	}
	_ = active.Discard()
}

func TestJanitorFailsClosedBeforeRemovalWhenAuthorityIsUnavailable(t *testing.T) {
	manager, _ := newManager(filepath.Join(t.TempDir(), "spool"), 64, 256)
	lease, _ := manager.Acquire(spoolIdentity(1))
	sealLease(t, lease)
	lookup := janitorLookup{err: errors.New("database unavailable")}
	if _, err := manager.reap(context.Background(), lookup,
		time.Now().UTC().Add(time.Second), 0); err == nil {
		t.Fatal("janitor accepted unavailable lease authority")
	}
	if _, err := os.Stat(lease.spoolPath); err != nil {
		t.Fatalf("spool removed without authority: %v", err)
	}
	_ = lease.Discard()
}

func sealLease(t *testing.T, lease *Lease) {
	t.Helper()
	writer, err := lease.OpenWriter()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("FCART1")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
}
