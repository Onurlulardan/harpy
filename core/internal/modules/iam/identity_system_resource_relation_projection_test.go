package iam

import (
	"errors"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

func TestIdentityRelationValuesOmitUnauthorizedTargets(t *testing.T) {
	reference := identityRelationReference{
		definition: identityRelationDefinition{
			field: "role_ids", target: publicresource.Roles, multiple: true,
		},
		ids: []uuid.UUID{
			uuid.MustParse("018f0f4e-7b61-7c25-8000-000000000096"),
		},
	}
	values, err := identityRelationValues(
		[]identityRelationReference{reference},
		map[publicresource.Resource]map[uuid.UUID]publicresource.RelationSummary{},
		map[publicresource.Resource]bool{publicresource.Roles: false},
	)
	if err != nil || len(values) != 0 {
		t.Fatalf("unauthorized values=%+v error=%v", values, err)
	}
}

func TestIdentityRelationValuesRejectMissingAuthorizedTarget(t *testing.T) {
	reference := identityRelationReference{
		definition: identityRelationDefinition{
			field: "parent_id", target: publicresource.Organizations,
		},
		ids: []uuid.UUID{
			uuid.MustParse("018f0f4e-7b61-7c25-8000-000000000097"),
		},
	}
	_, err := identityRelationValues(
		[]identityRelationReference{reference},
		map[publicresource.Resource]map[uuid.UUID]publicresource.RelationSummary{
			publicresource.Organizations: {},
		},
		map[publicresource.Resource]bool{publicresource.Organizations: true},
	)
	if !errors.Is(err, publicresource.ErrUnavailable) {
		t.Fatalf("missing authorized target error=%v", err)
	}
}
