package postgres

import (
	"time"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/retention"
	"github.com/google/uuid"
)

func successor(now time.Time, tenantID, jobID uuid.UUID,
	remaining bool,
) retentionapp.Successor {
	runAt := now.Add(time.Minute)
	if !remaining {
		runAt = time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, time.UTC)
		if !runAt.After(now) {
			runAt = runAt.Add(24 * time.Hour)
		}
	}
	reference, digest, _ := retentionapp.Reference(tenantID)
	return retentionapp.Successor{JobID: jobID, RunAt: runAt,
		PayloadReference: reference, PayloadSHA256: digest,
		DedupeKey: "notifications:retention:" + jobID.String(), Priority: 0}
}
