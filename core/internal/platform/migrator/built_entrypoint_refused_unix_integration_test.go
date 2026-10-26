//go:build integration && !windows

package migrator

import (
	"errors"
	"net"
	"os"
	"syscall"
	"testing"
)

func isBuiltEntrypointRefusedCause(err error) bool {
	return errors.Is(err, syscall.ECONNREFUSED)
}

func TestBuiltEntrypointConnectionRefusedClassification(t *testing.T) {
	refused := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &os.SyscallError{
			Syscall: "connect",
			Err:     syscall.ECONNREFUSED,
		},
	}
	if !isBuiltEntrypointConnectionRefused(refused) {
		t.Fatal("expected ECONNREFUSED dial error to be classified as refused")
	}

	refused.Net = "udp"
	if isBuiltEntrypointConnectionRefused(refused) {
		t.Fatal("non-TCP error must not be classified as a refused connection")
	}
}
