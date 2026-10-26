package documents

import (
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type APIAsyncPort struct {
	catalogs platformasync.WriteCatalogs
	jobs     jobsv1.Catalog
	clock    foundation.Clock
}

func NewAPIAsyncPort(catalogs platformasync.WriteCatalogs,
	clock foundation.Clock,
) (*APIAsyncPort, error) {
	jobs, _, err := catalogs.Snapshot()
	if err != nil || clock == nil {
		return nil, fmt.Errorf("Documents API async dependencies are required")
	}
	return &APIAsyncPort{catalogs: catalogs, jobs: jobs, clock: clock}, nil
}
