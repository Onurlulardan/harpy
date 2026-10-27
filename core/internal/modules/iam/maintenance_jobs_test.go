package iam

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func maintenanceKeyFixture(t *testing.T) MaintenanceDirectoryKey {
	t.Helper()
	material := [32]byte{1, 2, 3}
	id := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{22}, 12))
	key, err := NewMaintenanceDirectoryKey(id, material, true)
	if err != nil {
		t.Fatalf("new maintenance key: %v", err)
	}
	return key
}

func TestNewMaintenanceJobRuntimeConsumesInputOnEveryPath(t *testing.T) {
	valid := []MaintenanceDirectoryKey{maintenanceKeyFixture(t)}
	runtime, err := NewMaintenanceJobRuntime(valid)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	assertMaintenanceKeysWiped(t, valid)
	if runtime.Binding().Owner() != "iam" {
		t.Fatalf("binding owner = %q", runtime.Binding().Owner())
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close runtime: %v", err)
	}
	if runtime.Binding().Owner() != "" {
		t.Fatal("closed runtime exposed a binding")
	}

	invalid := []MaintenanceDirectoryKey{
		maintenanceKeyFixture(t), maintenanceKeyFixture(t),
	}
	if _, err := NewMaintenanceJobRuntime(invalid); err == nil {
		t.Fatal("duplicate runtime keys were accepted")
	}
	assertMaintenanceKeysWiped(t, invalid)
}

func assertMaintenanceKeysWiped(t *testing.T, keys []MaintenanceDirectoryKey) {
	t.Helper()
	for _, key := range keys {
		if key.id != "" || key.material != [32]byte{} || key.active {
			t.Fatalf("consumed key was not wiped: %#v", key)
		}
	}
}
