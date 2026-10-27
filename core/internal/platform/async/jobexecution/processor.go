package jobexecution

import (
	"context"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

// Processor owns the domain transaction and must durably finalize or defer the claim.
// Expected business failures are durable outcomes; returned errors are runtime failures.
type Processor interface {
	Policy() jobsv1.KindPolicy
	Process(context.Context, asyncpostgres.JobClaim) error
}
