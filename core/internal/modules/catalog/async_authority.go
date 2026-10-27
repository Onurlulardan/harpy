package catalog

import (
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func catalogJobAuthority(value platformtx.OwnerAuthority) jobsv1.Authority {
	return jobsv1.Authority{ScopeKind: jobsv1.ScopeTenant, ScopeID: value.ScopeID().String(),
		Domain: value.Domain(), CutoverGeneration: value.Generation(), WriterOrigin: jobsv1.WriterV2,
		GroupID: value.GroupID(), GroupGeneration: value.GroupGeneration()}
}

func catalogEventAuthority(value platformtx.OwnerAuthority) eventsv1.Authority {
	return eventsv1.Authority{ScopeKind: eventsv1.ScopeTenant, ScopeID: value.ScopeID().String(),
		Domain: value.Domain(), CutoverGeneration: value.Generation(), WriterOrigin: eventsv1.WriterV2,
		GroupID: value.GroupID(), GroupGeneration: value.GroupGeneration()}
}
