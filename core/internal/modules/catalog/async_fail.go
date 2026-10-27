package catalog

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (port *AsyncPort) FailBeforeEffect(
	ctx context.Context,
	scope platformtx.TenantScope,
	claim asyncpostgres.JobClaim,
	cause error,
) (jobsv1.State, error) {
	writer, _, err := port.bindAsync(ctx, scope)
	if err != nil {
		return "", err
	}
	message := "catalog resume failed"
	if cause != nil && cause.Error() != "" {
		message = cause.Error()
	}
	digest := sha256.Sum256([]byte(message))
	failure, err := asyncpostgres.NewJobFailure(
		jobsv1.ErrorTransient, "CATALOG_RESUME_TRANSIENT",
		hex.EncodeToString(digest[:]), time.Duration(0),
	)
	if err != nil {
		return "", asyncUnavailable()
	}
	result, err := writer.FailJobBeforeEffect(ctx, claim, failure)
	if err != nil {
		return "", err
	}
	return result.State(), nil
}
