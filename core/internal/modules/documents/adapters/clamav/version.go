package clamav

import (
	"bufio"
	"context"
	"errors"
	"strings"
	"time"

	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

func (scanner *Scanner) probeVersion(ctx context.Context) (string, time.Time, error) {
	connection, err := scanner.dialer.DialContext(ctx, "tcp", scanner.address)
	if err != nil {
		return "", time.Time{}, errors.Join(uploadapp.ErrScanUnavailable, err)
	}
	defer connection.Close()
	if err = connection.SetDeadline(scanner.clock.Now().Add(10 * time.Second)); err != nil {
		return "", time.Time{}, errors.Join(uploadapp.ErrScanUnavailable, err)
	}
	if _, err = connection.Write([]byte("zVERSION\x00")); err != nil {
		return "", time.Time{}, errors.Join(uploadapp.ErrScanUnavailable, err)
	}
	response, err := readTerminated(bufio.NewReader(connection))
	if err != nil {
		return "", time.Time{}, errors.Join(uploadapp.ErrScanUnavailable, err)
	}
	revision, at, err := parseVersion(response)
	if err != nil {
		return "", time.Time{}, err
	}
	now := scanner.clock.Now().UTC()
	if at.After(now.Add(5*time.Minute)) || now.Sub(at) > maximumScanAge {
		return "", time.Time{}, ErrStale
	}
	return revision, at, nil
}

func parseVersion(value string) (string, time.Time, error) {
	value = strings.TrimSpace(value)
	parts := strings.SplitN(value, "/", 3)
	if len(parts) != 3 || !strings.HasPrefix(parts[0], "ClamAV ") ||
		len(value) > 256 {
		return "", time.Time{}, ErrProtocol
	}
	databaseAt, err := time.ParseInLocation("Mon Jan _2 15:04:05 2006",
		strings.TrimSpace(parts[2]), time.UTC)
	if err != nil {
		return "", time.Time{}, ErrProtocol
	}
	return value, databaseAt.UTC(), nil
}

func readTerminated(reader *bufio.Reader) (string, error) {
	value, err := reader.ReadString(0)
	if err != nil || len(value) < 2 || len(value) > maximumResponse {
		return "", ErrProtocol
	}
	return strings.TrimSuffix(value, "\x00"), nil
}
