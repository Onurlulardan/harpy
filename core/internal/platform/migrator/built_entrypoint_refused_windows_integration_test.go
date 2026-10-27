//go:build integration && windows

package migrator

import (
	"errors"
	"net"
	"os"
	"testing"

	"golang.org/x/sys/windows"
)

func isBuiltEntrypointRefusedCause(err error) bool {
	return errors.Is(err, windows.WSAECONNREFUSED)
}

func TestBuiltEntrypointConnectionRefusedClassification(t *testing.T) {
	refused := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &os.SyscallError{
			Syscall: "connectex",
			Err:     windows.WSAECONNREFUSED,
		},
	}
	if !isBuiltEntrypointConnectionRefused(refused) {
		t.Fatal("expected WSAECONNREFUSED dial error to be classified as refused")
	}

	refused.Op = "read"
	if isBuiltEntrypointConnectionRefused(refused) {
		t.Fatal("non-dial error must not be classified as a refused connection")
	}
}
