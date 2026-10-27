package documents

import (
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func documentsJobAuthority(value platformtx.OwnerAuthority) jobsv1.Authority {
	return jobsv1.Authority{ScopeKind: jobsv1.ScopeTenant,
		ScopeID: value.ScopeID().String(), GroupID: value.GroupID(),
		GroupGeneration: value.GroupGeneration(), Domain: value.Domain(),
		CutoverGeneration: value.Generation(), WriterOrigin: jobsv1.WriterV2}
}
