package iam

import (
	"context"
	"fmt"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const identityAuthorityGroup = "iam_system_resources"

func (application *identityHTTPApplication) commandScopeRequest(
	ctx context.Context,
	evidence identityEvidenceSource,
	operationID string,
) (platformtx.ScopeRequest, error) {
	generations, err := application.pool.ResolveSnapshotDomainGenerations(
		ctx, evidence.Tenant(), []string{"iam", "system_resource_layout"},
	)
	if err != nil || len(generations) != 2 || generations["iam"] <= 0 ||
		generations["iam"] != generations["system_resource_layout"] {
		return platformtx.ScopeRequest{}, ErrIdentityHTTPUnavailable
	}
	request, err := platformtx.AdmitCommandScope(
		evidence.Subject(), evidence.Tenant(), operationID,
		identityAuthorityGroup, generations["iam"], generations,
	)
	if err != nil {
		return platformtx.ScopeRequest{}, ErrIdentityHTTPUnavailable
	}
	return request, nil
}

func (application *identityHTTPApplication) snapshotScopeRequest(
	ctx context.Context,
	evidence identityEvidenceSource,
	_ string,
) (platformtx.ScopeRequest, error) {
	request, err := iamHTTPSnapshotScopeRequest(ctx, application.pool, evidence)
	if err != nil {
		return platformtx.ScopeRequest{}, ErrIdentityHTTPUnavailable
	}
	return request, nil
}

func identityResourceKey(resource IdentityResource) (string, error) {
	switch resource {
	case IdentityUsers, IdentityRoles, IdentityOrganizations:
		return "iam." + string(resource), nil
	default:
		return "", fmt.Errorf("unknown IAM system resource")
	}
}
