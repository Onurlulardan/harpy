package billing

import (
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func subscriptionJobAuthority(authority platformtx.OwnerAuthority) jobsv1.Authority {
	return jobsv1.Authority{
		ScopeKind: jobsv1.ScopeTenant, ScopeID: authority.ScopeID().String(),
		Domain: authority.Domain(), CutoverGeneration: authority.Generation(),
		WriterOrigin: jobsv1.WriterV2, GroupID: authority.GroupID(),
		GroupGeneration: authority.GroupGeneration(),
	}
}

func billingEventAuthority(authority platformtx.OwnerAuthority) eventsv1.Authority {
	return eventsv1.Authority{
		ScopeKind: eventsv1.ScopeTenant, ScopeID: authority.ScopeID().String(),
		Domain: authority.Domain(), CutoverGeneration: authority.Generation(),
		WriterOrigin: eventsv1.WriterV2, GroupID: authority.GroupID(),
		GroupGeneration: authority.GroupGeneration(),
	}
}
