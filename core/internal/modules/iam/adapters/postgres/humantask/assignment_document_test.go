package humantask

import (
	"testing"
	"time"

	humantaskapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/humantask"
	"github.com/google/uuid"
)

func TestDecodeAssignmentPinsOnlyStableIdentityEvidence(t *testing.T) {
	tenant := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	organization := uuid.MustParse("10000000-0000-4000-8000-000000000002")
	request := humantaskapi.AssignmentRequest{
		TenantID: [16]byte(tenant), Assignment: humantaskapi.Candidate{
			Kind: "organization", ID: [16]byte(organization),
		}, VerifyAt: time.Date(2026, 7, 18, 15, 0, 0, 0, time.UTC),
	}
	document := []byte(`{"kind":"organization","id":"` + organization.String() +
		`","revision":7,"facts":[{"kind":"organization","id":"` +
		organization.String() + `","revision":7}]}`)
	evidence, err := decodeAssignment(document, request)
	if err != nil || !evidence.Matches(request) || evidence.Revision != 7 {
		t.Fatalf("assignment evidence = %+v, %v", evidence, err)
	}
}
