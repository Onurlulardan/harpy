package iam

import (
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestCurrentSessionScopeRequiresAtomicIAMLayoutGeneration(t *testing.T) {
	userID := uuid.MustParse("31000000-0000-4000-8000-000000000001")
	tenantID := uuid.MustParse("31000000-0000-4000-8000-000000000002")
	request, err := buildCurrentSessionScope(userID, tenantID, map[string]int64{
		"iam": 11, "system_resource_layout": 11,
	})
	if err != nil || request.Mode != platformtx.SnapshotQuery {
		t.Fatalf("request=%+v error=%v", request, err)
	}
	for name, generations := range map[string]map[string]int64{
		"single domain": {"iam": 11},
		"unequal":       {"iam": 11, "system_resource_layout": 12},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := buildCurrentSessionScope(
				userID, tenantID, generations,
			); err == nil {
				t.Fatal("unsafe current-session snapshot scope was admitted")
			}
		})
	}
}
