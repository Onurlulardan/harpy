package jobs

import (
	"bytes"
	"context"

	jobapp "github.com/Onurlulardan/harpy/core/internal/modules/work/application/jobs"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (store *Store) BeginJob(ctx context.Context, claim jobapp.Claim) (bool, error) {
	if !store.valid() || !claim.Valid() {
		return false, ports.ErrInvalid
	}
	affected, err := store.database.execute(ctx, beginJobSQL, claim.JobID, claim.Kind,
		claim.PayloadSHA256[:], claim.LeaseGeneration, claim.Now)
	if err != nil {
		return false, ports.ErrUnavailable
	}
	if affected == 1 {
		return false, nil
	}
	var kind, state string
	var payload []byte
	var generation int64
	if store.database.queryRow(ctx, loadJobSQL, claim.JobID).Scan(
		&kind, &payload, &generation, &state,
	) != nil || kind != claim.Kind || !bytes.Equal(payload, claim.PayloadSHA256[:]) ||
		generation != claim.LeaseGeneration || state != "SUCCEEDED" {
		return false, ports.ErrConflict
	}
	return true, nil
}

func (store *Store) SealJob(ctx context.Context, claim jobapp.Claim,
	digest [32]byte,
) error {
	if !store.valid() || !claim.Valid() || digest == [32]byte{} {
		return ports.ErrInvalid
	}
	return one(store.database.execute(ctx, completeJobSQL, claim.JobID, claim.Kind,
		claim.PayloadSHA256[:], claim.LeaseGeneration, digest[:], claim.Now))
}
