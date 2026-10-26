package workhttp

import (
	"testing"

	"github.com/google/uuid"
)

func TestAccessValidForJWTCompatibilityWithoutCohortHeaders(t *testing.T) {
	access, err := NewAccess(uuid.New(), uuid.New(), 1, 1,
		"work.calendar.list", "listLegacyCalendarEvents", "", "", true)
	if err != nil || !access.ValidFor("work.calendar.list", "listLegacyCalendarEvents", true) {
		t.Fatalf("jwt compatibility access = %#v err=%v", access, err)
	}
}
