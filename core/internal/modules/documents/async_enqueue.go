package documents

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (port *APIAsyncPort) enqueue(ctx context.Context, scope platformtx.TenantScope,
	jobID foundation.ID, kind, reference, digest, key string, runAt time.Time,
	payload []byte,
) error {
	if port == nil || !jobID.UUIDv7() || runAt.IsZero() || len(payload) == 0 ||
		sha256Hex(payload) != digest {
		return ErrDocumentsRuntime
	}
	writer, authority, err := port.bind(ctx, scope)
	if err != nil {
		return err
	}
	policy, err := port.jobs.Resolve(kind, 1)
	if err != nil || policy.OwnerDomain != "documents" {
		return ErrDocumentsRuntime
	}
	tenantID, effectKey := scope.TenantID().String(), key
	value, err := jobsv1.NewEnqueue(jobsv1.Enqueue{JobID: jobID.String(), Kind: kind,
		PayloadRevision: 1, TenantID: &tenantID,
		Authority: documentsJobAuthority(authority), PayloadReference: reference,
		PayloadSHA256: digest, RunAt: runAt.UTC().Format(time.RFC3339Nano),
		Priority: 100, DedupeKey: key, EffectIdempotencyKey: &effectKey,
		MaximumAttempts: policy.MaximumAttempts}, port.jobs)
	if err != nil {
		return ErrDocumentsRuntime
	}
	result, err := writer.EnqueueJob(ctx, value)
	if err != nil || result.ID != jobID.String() {
		return ErrDocumentsRuntime
	}
	return nil
}

func sha256Hex(document []byte) string {
	digest := sha256.Sum256(document)
	return hex.EncodeToString(digest[:])
}

var _ asyncpostgres.JobWriter = (asyncpostgres.Writer)(nil)
