package iam

import (
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type systemTransferExportReconciliationBackoff struct {
	jitter platformasync.JitterSource
}

func newSystemTransferExportReconciliationBackoff() systemTransferExportReconciliationBackoff {
	return systemTransferExportReconciliationBackoff{jitter: systemTransferImportJitter{}}
}

func (backoff systemTransferExportReconciliationBackoff) RetryDelay(
	attempt int,
) (time.Duration, error) {
	return platformasync.RetryDelay(attempt, backoff.jitter)
}
