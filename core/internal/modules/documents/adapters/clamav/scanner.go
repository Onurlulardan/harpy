package clamav

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

var (
	ErrConfiguration = errors.New("documents ClamAV configuration is invalid")
	ErrProtocol      = errors.New("documents ClamAV protocol failed")
	ErrStale         = errors.New("documents ClamAV signatures are stale")
)

const (
	maximumFrame    = 1024 * 1024
	maximumResponse = 512
	maximumScanAge  = 24 * time.Hour
)

type Dialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}

type Clock interface{ Now() time.Time }

type Scanner struct {
	address string
	dialer  Dialer
	clock   Clock
}

func NewVerified(ctx context.Context, address string, dialer Dialer,
	clock Clock,
) (*Scanner, error) {
	if ctx == nil || ctx.Err() != nil || dialer == nil || clock == nil ||
		strings.TrimSpace(address) != address {
		return nil, ErrConfiguration
	}
	if _, _, err := net.SplitHostPort(address); err != nil {
		return nil, ErrConfiguration
	}
	scanner := &Scanner{address: address, dialer: dialer, clock: clock}
	if _, _, err := scanner.probeVersion(ctx); err != nil {
		return nil, err
	}
	return scanner, nil
}

func (scanner *Scanner) Open(ctx context.Context, filename, declaredMIME string,
	expectedBytes int64,
) (uploadapp.ScanSession, error) {
	if scanner == nil || ctx == nil || ctx.Err() != nil || filename == "" ||
		declaredMIME == "" || expectedBytes < 1 {
		return nil, uploadapp.ErrInvalid
	}
	revision, signatureAt, err := scanner.probeVersion(ctx)
	if err != nil {
		return nil, err
	}
	connection, err := scanner.dialer.DialContext(ctx, "tcp", scanner.address)
	if err != nil {
		return nil, errors.Join(uploadapp.ErrScanUnavailable, err)
	}
	if err = connection.SetDeadline(scanner.clock.Now().Add(5 * time.Minute)); err != nil {
		_ = connection.Close()
		return nil, errors.Join(uploadapp.ErrScanUnavailable, err)
	}
	if _, err = connection.Write([]byte("zINSTREAM\x00")); err != nil {
		_ = connection.Close()
		return nil, errors.Join(uploadapp.ErrScanUnavailable, err)
	}
	return &session{connection: connection, filename: filename,
		declaredMIME: declaredMIME, expectedBytes: expectedBytes,
		revision: revision, signatureAt: signatureAt, clock: scanner.clock,
		prefix: make([]byte, 0, maximumFrame)}, nil
}

var _ uploadapp.Scanner = (*Scanner)(nil)
