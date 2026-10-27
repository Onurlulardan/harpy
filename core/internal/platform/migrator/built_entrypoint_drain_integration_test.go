//go:build integration

package migrator

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

func openBuiltEntrypointDrainConnection(t *testing.T, address string) net.Conn {
	t.Helper()
	connection, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		t.Fatalf("open built core-api drain connection: %v", err)
	}
	if err := connection.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		_ = connection.Close()
		t.Fatal(err)
	}
	partialRequest := "GET /held-drain HTTP/1.1\r\nHost: " + address +
		"\r\nX-Correlation-ID: built-held-drain\r\n"
	if _, err := io.WriteString(connection, partialRequest); err != nil {
		_ = connection.Close()
		t.Fatalf("write built core-api drain connection: %v", err)
	}
	if err := connection.SetWriteDeadline(time.Time{}); err != nil {
		_ = connection.Close()
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	return connection
}

func assertBuiltEntrypointDrainClosesListener(
	t *testing.T,
	address string,
	held net.Conn,
	exited <-chan struct{},
) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		probe, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if isBuiltEntrypointConnectionRefused(err) {
			assertBuiltEntrypointStillDraining(t, exited)
			if err := held.Close(); err != nil {
				t.Fatalf("release built core-api drain connection: %v", err)
			}
			return
		}
		if err == nil {
			_ = probe.Close()
		}
		assertBuiltEntrypointStillDraining(t, exited)
		if time.Now().After(deadline) {
			t.Fatalf("built core-api listener did not refuse during drain: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func isBuiltEntrypointConnectionRefused(err error) bool {
	var operationError *net.OpError
	if !errors.As(err, &operationError) {
		return false
	}
	if operationError.Op != "dial" || operationError.Net != "tcp" {
		return false
	}
	return isBuiltEntrypointRefusedCause(operationError)
}

func assertBuiltEntrypointStillDraining(t *testing.T, exited <-chan struct{}) {
	t.Helper()
	select {
	case <-exited:
		t.Fatal("built core-api exited before held drain request was released")
	default:
	}
}

func assertBuiltEntrypointPublicClosed(t *testing.T, address string) {
	t.Helper()
	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatalf("built core-api public listener remained bound after drain: %v", err)
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
}
