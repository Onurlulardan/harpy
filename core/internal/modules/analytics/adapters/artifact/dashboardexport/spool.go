package dashboardexportartifact

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	spoolapi "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactspool/api"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

type reservedSpool struct {
	identity exportapp.SpoolReservationIdentity
	lease    *spoolapi.Lease
}

func (spool *reservedSpool) Identity() exportapp.SpoolReservationIdentity {
	if spool == nil {
		return exportapp.SpoolReservationIdentity{}
	}
	return spool.identity
}

func (spool *reservedSpool) OpenWriter() (io.WriteCloser, error) {
	if spool == nil || spool.lease == nil {
		return nil, exportapp.ErrUnavailable
	}
	return spool.lease.OpenWriter()
}

func (spool *reservedSpool) OpenReader() (io.ReadCloser, error) {
	if spool == nil || spool.lease == nil {
		return nil, exportapp.ErrUnavailable
	}
	return spool.lease.OpenReader()
}

func (spool *reservedSpool) Discard() error {
	if spool == nil || spool.lease == nil {
		return nil
	}
	err := spool.lease.Discard()
	if err == nil {
		spool.lease = nil
	}
	return err
}

func (runtime *Runtime) Reserve(
	identity exportapp.SpoolReservationIdentity,
) (exportapp.EncryptedSpool, error) {
	if runtime == nil || identity.ExportID.Version() != 7 ||
		identity.AttemptGeneration < 1 || identity.AttemptGeneration > exportapp.MaximumAttempts ||
		identity.LeaseOwnerID == [16]byte{} {
		return nil, exportapp.ErrInvalid
	}
	lease, err := runtime.spool.Acquire(spoolapi.ReservationIdentity{
		ExportID: identity.ExportID, AttemptGeneration: identity.AttemptGeneration,
		LeaseOwnerID: identity.LeaseOwnerID})
	if errors.Is(err, spoolapi.ErrCapacity) {
		return nil, exportapp.ErrSpoolCapacity
	}
	if err != nil {
		return nil, errors.Join(exportapp.ErrUnavailable, err)
	}
	return &reservedSpool{identity: identity, lease: lease}, nil
}

func (runtime *Runtime) Reap(ctx context.Context, lookup exportapp.SpoolLeaseLookup,
	now time.Time,
) error {
	if runtime == nil || lookup == nil || ctx == nil || ctx.Err() != nil || now.IsZero() {
		return exportapp.ErrInvalid
	}
	_, err := runtime.spool.Reap(ctx, activeSpoolLookup{next: lookup}, now)
	if err != nil {
		return fmt.Errorf("reap Analytics dashboard export spool: %w", err)
	}
	return nil
}

type activeSpoolLookup struct{ next exportapp.SpoolLeaseLookup }

func (lookup activeSpoolLookup) Active(ctx context.Context,
	identity spoolapi.ReservationIdentity,
) (bool, error) {
	return lookup.next.ActiveSpoolReservation(ctx, exportapp.SpoolReservationIdentity{
		ExportID: identity.ExportID, AttemptGeneration: identity.AttemptGeneration,
		LeaseOwnerID: identity.LeaseOwnerID})
}

var _ exportapp.SpoolAllocator = (*Runtime)(nil)
