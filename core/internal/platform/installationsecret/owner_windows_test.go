//go:build windows

package installationsecret

import (
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestReadRejectsBroadWindowsACL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "initial-password")
	if err := Publish(path, canonicalWire(0x5d)); err != nil {
		t.Fatal(err)
	}
	descriptor, err := windows.SecurityDescriptorFromString("D:P(A;;FA;;;WD)")
	if err != nil {
		t.Fatal(err)
	}
	dacl, _, err := descriptor.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err = windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil, nil, dacl, nil); err != nil {
		t.Fatal(err)
	}
	if err = Read(path, func([]byte) error { return nil }); err != ErrFenced {
		t.Fatalf("broad ACL err=%v", err)
	}
}
