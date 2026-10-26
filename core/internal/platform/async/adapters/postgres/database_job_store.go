package postgres

import (
	"context"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type DatabaseJobStore interface {
	ProcessDatabaseJob(
		context.Context,
		JobClaim,
		jobsv1.KindPolicy,
		platformasync.DatabaseJobBinding,
	) error
}
