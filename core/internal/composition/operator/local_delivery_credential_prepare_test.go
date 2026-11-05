package operator

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
	"github.com/Onurlulardan/harpy/core/internal/platform/installationsecret"
)

func TestPrepareInitialCredentialCreatesOnlyForUnstagedDelivery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "initial-admin-password")
	if err := prepareInitialCredential(path, tenancy.InstallationDelivery{},
		installationdelivery.ErrNotFound); err != nil {
		t.Fatal(err)
	}
	var first []byte
	if err := installationsecret.Read(path, func(password []byte) error {
		first = append([]byte(nil), password...)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(first) != 32 {
		t.Fatalf("password length=%d", len(first))
	}
	defer clear(first)
	if err := prepareInitialCredential(path, tenancy.InstallationDelivery{},
		installationdelivery.ErrNotFound); err != nil {
		t.Fatal(err)
	}
	var second []byte
	if err := installationsecret.Read(path, func(password []byte) error {
		second = append([]byte(nil), password...)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	defer clear(second)
	if !bytes.Equal(first, second) {
		t.Fatal("existing initial credential was replaced")
	}
}

func TestPrepareInitialCredentialPropagatesUnavailableWithoutCreatingPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "initial-admin-password")
	err := prepareInitialCredential(path, tenancy.InstallationDelivery{},
		installationdelivery.ErrUnavailable)
	if !errors.Is(err, installationdelivery.ErrUnavailable) {
		t.Fatalf("err=%v", err)
	}
	if _, statErr := os.Lstat(path); !os.IsNotExist(statErr) {
		t.Fatalf("password was created after unavailable read: %v", statErr)
	}
}

func TestPrepareInitialCredentialDoesNotRecreateMissingStagedPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "initial-admin-password")
	err := prepareInitialCredential(path,
		tenancy.InstallationDelivery{State: tenancy.InstallationDeliveryStateReady}, nil)
	if err == nil {
		t.Fatal("missing password for a staged delivery was accepted")
	}
	if _, statErr := os.Lstat(path); !os.IsNotExist(statErr) {
		t.Fatalf("password was regenerated for a staged delivery: %v", statErr)
	}
}

func TestPrepareDeliveryRejectsUnstagedAndUnavailableRecords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "initial-admin-password")
	for _, readErr := range []error{
		installationdelivery.ErrNotFound,
		installationdelivery.ErrUnavailable,
	} {
		if err := prepareDeliveryCredential(path, tenancy.InstallationDelivery{}, readErr); !errors.Is(err, readErr) {
			t.Fatalf("read error=%v returned=%v", readErr, err)
		}
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("delivery prepare created a password: %v", err)
	}
}

func TestPrepareDeliveryRequiresExistingPasswordForReadyState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "initial-admin-password")
	err := prepareDeliveryCredential(path,
		tenancy.InstallationDelivery{State: tenancy.InstallationDeliveryStateReady}, nil)
	if err == nil {
		t.Fatal("ready delivery without its original password was accepted")
	}
	if _, statErr := os.Lstat(path); !os.IsNotExist(statErr) {
		t.Fatalf("delivery prepare generated a replacement password: %v", statErr)
	}
}

func TestPrepareDeliveryFencesPasswordAfterSecretDeletion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "initial-admin-password")
	for _, state := range []tenancy.InstallationDeliveryState{
		tenancy.InstallationDeliveryStateSecretDeleted,
		tenancy.InstallationDeliveryStateCompleted,
	} {
		if err := prepareDeliveryCredential(path,
			tenancy.InstallationDelivery{State: state}, nil); err != nil {
			t.Fatalf("state=%s err=%v", state, err)
		}
	}
	if err := os.WriteFile(path, []byte("unexpected"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, state := range []tenancy.InstallationDeliveryState{
		tenancy.InstallationDeliveryStateSecretDeleted,
		tenancy.InstallationDeliveryStateCompleted,
	} {
		if err := prepareDeliveryCredential(path,
			tenancy.InstallationDelivery{State: state}, nil); !errors.Is(err, installationsecret.ErrFenced) {
			t.Fatalf("state=%s err=%v", state, err)
		}
	}
}

func TestPrepareInitialCredentialDoesNotCreateAfterSecretDeletion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "initial-admin-password")
	for _, state := range []tenancy.InstallationDeliveryState{
		tenancy.InstallationDeliveryStateSecretDeleted,
		tenancy.InstallationDeliveryStateCompleted,
	} {
		if err := prepareInitialCredential(path,
			tenancy.InstallationDelivery{State: state}, nil); err != nil {
			t.Fatalf("state=%s err=%v", state, err)
		}
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("password was recreated after deletion: %v", err)
	}
}
