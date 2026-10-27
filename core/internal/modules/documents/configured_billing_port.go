package documents

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type ConfiguredDocumentsBilling interface {
	ReserveDocumentsActor(context.Context, platformtx.TenantScope,
		string, int64, time.Time) (string, bool, bool, error)
	StartDocumentsActor(context.Context, platformtx.TenantScope, string, string) error
	CommitDocumentsActor(context.Context, platformtx.TenantScope,
		string, string, int64) error
	ReleaseDocumentsActor(context.Context, platformtx.TenantScope, string, string) error
	ReserveDocumentsJob(context.Context, platformtx.JobTenantScope,
		string, string, int64, time.Time) (bool, error)
	StartDocumentsJob(context.Context, platformtx.JobTenantScope, string, string) error
	CommitDocumentsJob(context.Context, platformtx.JobTenantScope,
		string, string, int64) error
	ReleaseDocumentsJob(context.Context, platformtx.JobTenantScope, string, string) error
	DecreaseDocumentsJob(context.Context, platformtx.JobTenantScope,
		string, int64, string) error
	KeepDocumentsJob(context.Context, platformtx.JobTenantScope, string, string) error
}
