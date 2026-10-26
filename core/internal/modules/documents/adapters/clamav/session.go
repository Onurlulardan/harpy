package clamav

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"net"
	"strings"
	"time"

	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

type session struct {
	connection             net.Conn
	filename, declaredMIME string
	expectedBytes, written int64
	revision               string
	signatureAt            time.Time
	clock                  Clock
	prefix                 []byte
	finished               bool
}

func (stream *session) Write(value []byte) (int, error) {
	if stream == nil || stream.finished || stream.connection == nil ||
		stream.written+int64(len(value)) > stream.expectedBytes {
		return 0, ErrProtocol
	}
	if remaining := maximumFrame - len(stream.prefix); remaining > 0 {
		count := len(value)
		if count > remaining {
			count = remaining
		}
		stream.prefix = append(stream.prefix, value[:count]...)
	}
	written := 0
	for written < len(value) {
		count := len(value) - written
		if count > maximumFrame {
			count = maximumFrame
		}
		frame := make([]byte, 4)
		binary.BigEndian.PutUint32(frame, uint32(count))
		if err := writeFull(stream.connection, frame); err != nil {
			return written, err
		}
		if err := writeFull(stream.connection, value[written:written+count]); err != nil {
			return written, err
		}
		written += count
	}
	stream.written += int64(written)
	return written, nil
}

func (stream *session) Finish(ctx context.Context) (uploadapp.ScanResult, error) {
	if stream == nil || stream.finished || ctx == nil || ctx.Err() != nil ||
		stream.written != stream.expectedBytes {
		return uploadapp.ScanResult{}, ErrProtocol
	}
	stream.finished = true
	defer stream.connection.Close()
	if err := writeFull(stream.connection, []byte{0, 0, 0, 0}); err != nil {
		return uploadapp.ScanResult{}, errors.Join(uploadapp.ErrScanUnavailable, err)
	}
	response, err := readTerminated(bufio.NewReader(stream.connection))
	if err != nil {
		return uploadapp.ScanResult{}, errors.Join(uploadapp.ErrScanUnavailable, err)
	}
	kind, detectedMIME, err := detect(stream.filename, stream.declaredMIME, stream.prefix)
	if err != nil {
		return uploadapp.ScanResult{}, err
	}
	result := uploadapp.ScanResult{DetectedKind: kind, DetectedMIME: detectedMIME,
		SignatureRevision: stream.revision, ScannedAt: stream.clock.Now().UTC()}
	switch {
	case strings.HasSuffix(response, " OK"):
		result.Clean = true
		return result, nil
	case strings.HasSuffix(response, " FOUND"):
		return result, nil
	default:
		return uploadapp.ScanResult{}, errors.Join(uploadapp.ErrScanUnavailable, ErrProtocol)
	}
}

func writeFull(connection net.Conn, value []byte) error {
	for len(value) > 0 {
		written, err := connection.Write(value)
		if err != nil {
			return err
		}
		value = value[written:]
	}
	return nil
}

var _ uploadapp.ScanSession = (*session)(nil)
