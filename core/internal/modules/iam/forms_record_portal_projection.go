package iam

import (
	"context"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *FormsRecordAccessRuntime) ProjectPortalSubjectLinks(
	ctx context.Context, input PortalSubjectLinkProjectionInput,
) (map[uuid.UUID]PortalSubjectLink, error) {
	if runtime == nil || runtime.pool == nil || runtime.coordinator == nil ||
		runtime.portalUsers == nil || ctx == nil || ctx.Err() != nil || !input.valid() {
		return nil, ErrFormsRecordAccessUnavailable
	}
	generations, err := runtime.pool.ResolveSnapshotDomainGenerations(
		ctx, input.TenantID, []string{"iam", "system_resource_layout"},
	)
	if err != nil || !validIAMLayoutSnapshotGenerations(generations) {
		return nil, ErrFormsRecordAccessUnavailable
	}
	request, err := queryapp.SnapshotScopeRequest(input.PrincipalID, input.TenantID,
		map[string]int64{"iam": generations["iam"]})
	if err != nil {
		return nil, ErrFormsRecordAccessUnavailable
	}
	var links []portaluser.SubjectLink
	err = runtime.coordinator.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var projectErr error
			links, projectErr = runtime.portalUsers.ProjectSubjectLinks(scoped, scope,
				portaluser.ProjectionAccess{PrincipalID: input.PrincipalID, TenantID: input.TenantID,
					TokenVersion: input.TokenVersion, AuthorizationVersion: input.AuthorizationVersion,
					AuthorizationClock: input.Clock},
				input.FormID, input.FormVersionID, input.RecordIDs)
			return projectErr
		})
	if err != nil {
		return nil, ErrFormsRecordAccessUnavailable
	}
	result := make(map[uuid.UUID]PortalSubjectLink, len(links))
	for _, link := range links {
		projected := PortalSubjectLink{ID: link.UserID, Status: link.Status,
			AccountType: link.AccountType, Revision: link.Revision}
		if !projected.Valid() {
			return nil, ErrFormsRecordAccessUnavailable
		}
		result[link.SubjectRecordID] = projected
	}
	return result, nil
}
