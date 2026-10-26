package iam

import (
	"encoding/json"
	"errors"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

func TestIdentityRelationSummaryUsesHumanReadableResourceIdentity(t *testing.T) {
	id := uuid.MustParse("018f0f4e-7b61-7c25-8000-000000000094")
	userItem := json.RawMessage(`{
		"id":"` + id.String() + `",
		"display_name":"Local Instance Owner",
		"email":"admin@example.com",
		"status":"active"
	}`)
	actualID, summary, err := identityRelationSummary(
		publicresource.Users, userItem,
		map[string]struct{}{
			"id": {}, "display_name": {}, "email": {}, "status": {},
		},
	)
	summaryID, display, details, summaryErr := summary.Values()
	if err != nil || summaryErr != nil || actualID != id ||
		summaryID.String() != id.String() ||
		display != "Local Instance Owner" ||
		details["email"] != "admin@example.com" ||
		details["status"] != "active" {
		t.Fatalf("id=%s display=%q details=%v errors=%v/%v",
			actualID, display, details, err, summaryErr)
	}
}

func TestIdentityRelationSummaryFailsClosedOnUnknownOrTechnicalDisplay(
	t *testing.T,
) {
	id := "018f0f4e-7b61-7c25-8000-000000000095"
	for _, item := range []json.RawMessage{
		json.RawMessage(`{"id":"` + id + `","name":"Role","secret":"leak"}`),
		json.RawMessage(`{"id":"` + id + `","name":"` + id + `"}`),
		json.RawMessage(`{"id":"` + id + `","display_name":null}`),
	} {
		_, _, err := identityRelationSummary(
			publicresource.Roles, item,
			map[string]struct{}{"id": {}, "name": {}},
		)
		if !errors.Is(err, publicresource.ErrUnavailable) {
			t.Fatalf("item=%s error=%v", item, err)
		}
	}
}
