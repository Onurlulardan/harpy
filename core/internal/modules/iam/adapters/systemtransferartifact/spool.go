package systemtransferartifact

import (
	"errors"
	"io"
	"sync"

	spoolapi "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactspool/api"
	"github.com/google/uuid"
)

func (runtime *Runtime) acquireSpool(artifactID uuid.UUID, generation int64,
	leaseOwner uuid.UUID, stable bool) (*spoolapi.Lease, error) {
	if runtime == nil || runtime.spools == nil || artifactID == uuid.Nil ||
		generation < 1 || generation > 3 || leaseOwner == uuid.Nil {
		return nil, errInvalid
	}
	reservationID := artifactID
	if !stable {
		var err error
		reservationID, err = uuid.NewV7()
		if err != nil {
			return nil, errors.Join(errUnavailable, err)
		}
	}
	lease, err := runtime.spools.Acquire(spoolapi.ReservationIdentity{
		ExportID: reservationID, AttemptGeneration: generation,
		LeaseOwnerID: leaseOwner})
	if err != nil {
		return nil, errors.Join(errUnavailable, err)
	}
	return lease, nil
}

type discardingReader struct {
	reader io.ReadCloser
	lease  *spoolapi.Lease
	once   sync.Once
	err    error
}

func (reader *discardingReader) Read(value []byte) (int, error) {
	if reader == nil || reader.reader == nil {
		return 0, errUnavailable
	}
	return reader.reader.Read(value)
}

func (reader *discardingReader) Close() error {
	if reader == nil {
		return errUnavailable
	}
	reader.once.Do(func() {
		reader.err = errors.Join(reader.reader.Close(), reader.lease.Discard())
		reader.reader = nil
		reader.lease = nil
	})
	return reader.err
}

func openDiscardingReader(lease *spoolapi.Lease) (io.ReadCloser, error) {
	if lease == nil {
		return nil, errUnavailable
	}
	reader, err := lease.OpenReader()
	if err != nil {
		return nil, errors.Join(errUnavailable, err, lease.Discard())
	}
	return &discardingReader{reader: reader, lease: lease}, nil
}
