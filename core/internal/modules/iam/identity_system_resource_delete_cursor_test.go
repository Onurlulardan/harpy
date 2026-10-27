package iam

import (
	"errors"
	"testing"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	"github.com/google/uuid"
)

func TestIdentityUserDeleteRelationQueriesFailClosedOutsideBounds(t *testing.T) {
	oversized := make(
		[]identityfacade.Assignment, identityUserDeleteRelationBatchSize+1,
	)
	if _, _, err := identityUserDeleteAssignmentCursor(
		identityfacade.IdentityID{}, oversized,
	); !errors.Is(err, ErrIdentityHTTPUnavailable) {
		t.Fatalf("oversized relation page error = %v", err)
	}
	after := identityfacade.IdentityID(
		uuid.MustParse("30000000-0000-4000-8000-000000000002"),
	)
	stale := []identityfacade.Membership{{ID: after}}
	if _, _, err := identityUserDeleteMembershipCursor(
		after, stale,
	); !errors.Is(err, ErrIdentityHTTPUnavailable) {
		t.Fatalf("stale relation cursor error = %v", err)
	}
}
