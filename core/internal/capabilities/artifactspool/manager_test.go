package artifactspool

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"
)

func TestManagerEnforcesCrossProcessSlotCapacityAndReleasesExactlyOnce(t *testing.T) {
	root := filepath.Join(t.TempDir(), "spool")
	first, err := newManager(root, 64, 256)
	if err != nil {
		t.Fatal(err)
	}
	second, err := newManager(root, 64, 256)
	if err != nil {
		t.Fatal(err)
	}
	leases := make([]*Lease, 0, 4)
	for index := 0; index < 4; index++ {
		lease, err := first.Acquire(spoolIdentity(index + 1))
		if err != nil {
			t.Fatal(err)
		}
		leases = append(leases, lease)
	}
	if _, err := second.Acquire(spoolIdentity(1)); !errors.Is(err, ErrAlreadyReserved) {
		t.Fatalf("duplicate reservation error = %v", err)
	}
	if _, err := second.Acquire(spoolIdentity(5)); !errors.Is(err, ErrCapacity) {
		t.Fatalf("fifth reservation error = %v", err)
	}
	if err := leases[1].Discard(); err != nil {
		t.Fatal(err)
	}
	replacement, err := second.Acquire(spoolIdentity(6))
	if err != nil {
		t.Fatal(err)
	}
	if err := replacement.Discard(); err != nil {
		t.Fatal(err)
	}
	for _, lease := range leases {
		if err := lease.Discard(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLeaseSealsExactCiphertextAndRejectsOverflow(t *testing.T) {
	manager, err := newManager(filepath.Join(t.TempDir(), "spool"), 64, 256)
	if err != nil {
		t.Fatal(err)
	}
	lease, err := manager.Acquire(spoolIdentity(1))
	if err != nil {
		t.Fatal(err)
	}
	writer, err := lease.OpenWriter()
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("FCART1-encrypted-bytes")
	if _, err := writer.Write(want); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(make([]byte, 64)); err == nil {
		t.Fatal("reservation overflow accepted")
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	reader, err := lease.OpenReader()
	if err != nil {
		t.Fatal(err)
	}
	seen, err := io.ReadAll(reader)
	_ = reader.Close()
	if err != nil || string(seen) != string(want) {
		t.Fatalf("spool = %q, %v", seen, err)
	}
	info, err := os.Stat(lease.spoolPath)
	permissionsUnsafe := runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0
	if err != nil || info.Size() != int64(len(want)) || permissionsUnsafe {
		t.Fatalf("spool info = %#v, %v", info, err)
	}
	if err := lease.Discard(); err != nil {
		t.Fatal(err)
	}
}

func spoolIdentity(marker int) ReservationIdentity {
	value := [16]byte{0: 1, 6: 0x70, 8: 0x80, 15: byte(marker)}
	return ReservationIdentity{
		ExportID:          uuid.UUID(value),
		AttemptGeneration: int64((marker-1)%3 + 1),
		LeaseOwnerID:      uuid.MustParse("22222222-2222-4222-8222-222222222222"),
	}
}
