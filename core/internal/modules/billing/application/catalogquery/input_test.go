package catalogquery

import (
	"errors"
	"testing"
	"time"
)

func TestRequestsRequireCanonicalCursorLimitKeyAndInstant(t *testing.T) {
	request, err := BuildOfferingListRequest(25, "", fixtureInstant)
	if err != nil || request.Limit() != 25 || !request.Instant().Equal(fixtureInstant) {
		t.Fatalf("valid list request = %#v, %v", request, err)
	}
	cursor := encodeCursor("starter")
	request, err = BuildOfferingListRequest(25, cursor, fixtureInstant)
	if err != nil || request.after != "starter" {
		t.Fatalf("cursor request = %#v, %v", request, err)
	}
	invalid := []struct {
		limit  int
		cursor string
		at     time.Time
	}{{0, "", fixtureInstant}, {MaximumPageSize + 1, "", fixtureInstant},
		{1, "not+canonical", fixtureInstant},
		{1, encodeCursor("A"), fixtureInstant},
		{1, "", fixtureInstant.In(time.FixedZone("other", 0))},
		{1, "", fixtureInstant.Add(time.Nanosecond)}}
	for _, value := range invalid {
		if _, err = BuildOfferingListRequest(value.limit, value.cursor, value.at); !errors.Is(err, ErrRequest) {
			t.Fatalf("invalid list request accepted: %#v, %v", value, err)
		}
	}
	if _, err = BuildOfferingRequest("A", fixtureInstant); !errors.Is(err, ErrRequest) {
		t.Fatalf("invalid key accepted: %v", err)
	}
	if _, err = NewActiveCampaignRequest(time.Time{}); !errors.Is(err, ErrRequest) {
		t.Fatalf("invalid campaign instant accepted: %v", err)
	}
}
