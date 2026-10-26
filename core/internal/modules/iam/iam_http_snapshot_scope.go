package iam

import (
	"context"
	"errors"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var errIAMHTTPSnapshotScope = errors.New("IAM HTTP snapshot scope is unavailable")

func iamHTTPSnapshotScopeRequest(
	ctx context.Context,
	pool IdentityHTTPPlatform,
	evidence identityEvidenceSource,
) (platformtx.ScopeRequest, error) {
	generations, err := pool.ResolveSnapshotDomainGenerations(
		ctx, evidence.Tenant(), []string{"iam", "system_resource_layout"},
	)
	if err != nil || !validIAMLayoutSnapshotGenerations(generations) {
		return platformtx.ScopeRequest{}, errIAMHTTPSnapshotScope
	}
	request, err := queryapp.SnapshotScopeRequest(
		evidence.Subject(), evidence.Tenant(), map[string]int64{
			"iam": generations["iam"],
		},
	)
	if err != nil {
		return platformtx.ScopeRequest{}, errIAMHTTPSnapshotScope
	}
	return request, nil
}

func validIAMLayoutSnapshotGenerations(generations map[string]int64) bool {
	return len(generations) == 2 && generations["iam"] > 0 &&
		generations["iam"] == generations["system_resource_layout"]
}
