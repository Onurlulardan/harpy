package work

import (
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestWorkHTTPScopeUsesDottedAuthority(t *testing.T) {
	access, err := NewHTTPAccess(uuid.New(), uuid.New(), 1, 1,
		"work.calendar.list", "listLegacyCalendarEvents", "frontend-v3", "dev", true)
	if err != nil {
		t.Fatal(err)
	}
	if access.Authority() != "work.calendar.list" {
		t.Fatal("HTTP snapshot/command must use dotted authority")
	}
	_, err = platformtx.NewSnapshotAuthorityIssuer().Issue(
		"listLegacyCalendarEvents", map[string]int64{"work": 1})
	if err == nil {
		t.Fatal("camelCase OpenAPI operation id must not admit a snapshot")
	}
	if _, err = platformtx.NewSnapshotAuthorityIssuer().Issue(
		access.Authority(), map[string]int64{"work": 1}); err != nil {
		t.Fatal(err)
	}
	_, err = platformtx.AdmitCommandScope(uuid.New(), uuid.New(),
		"createLegacyCalendarEvent", "work", 1, map[string]int64{"work": 1})
	if err == nil {
		t.Fatal("camelCase OpenAPI operation id must not admit a command")
	}
	if _, err = platformtx.AdmitCommandScope(uuid.New(), uuid.New(),
		"work.calendar.change", "work", 1, map[string]int64{"work": 1}); err != nil {
		t.Fatal(err)
	}
}
