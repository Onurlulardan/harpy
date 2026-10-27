package documents

import (
	"context"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

type jobAuthorityContextKey struct{}

func withJobAuthority(ctx context.Context,
	authority effectsv1.Authority,
) context.Context {
	value := jobsv1.Authority{ScopeKind: jobsv1.ScopeKind(authority.ScopeKind),
		ScopeID: authority.ScopeID, GroupID: authority.GroupID,
		GroupGeneration: authority.GroupGeneration, Domain: authority.Domain,
		CutoverGeneration: authority.CutoverGeneration,
		WriterOrigin:      jobsv1.WriterOrigin(authority.WriterOrigin)}
	return context.WithValue(ctx, jobAuthorityContextKey{}, value)
}

func jobAuthorityFromContext(ctx context.Context) (jobsv1.Authority, bool) {
	if ctx == nil {
		return jobsv1.Authority{}, false
	}
	authority, found := ctx.Value(jobAuthorityContextKey{}).(jobsv1.Authority)
	return authority, found && authority.Validate() == nil &&
		authority.ScopeKind == jobsv1.ScopeTenant &&
		authority.WriterOrigin == jobsv1.WriterV2 && authority.Domain == "documents"
}
