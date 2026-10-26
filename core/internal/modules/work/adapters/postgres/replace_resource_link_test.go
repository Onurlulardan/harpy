package postgres

import (
	"strings"
	"testing"
)

func TestReplaceResourceLinkReleasesOnlyActiveTaskLinks(t *testing.T) {
	for _, fragment := range []string{
		"work_resource_kind = 'TASK'", "work_resource_id = $1",
		"state = 'ACTIVE'", "state = 'RELEASED'", "row_version = row_version + 1",
	} {
		if !strings.Contains(releaseTaskResourceLinksSQL, fragment) {
			t.Fatalf("resource replacement SQL missing %q", fragment)
		}
	}
}

func TestResourceLinkLockLoadIsBoundedAndCanonical(t *testing.T) {
	for _, fragment := range []string{
		"state = 'ACTIVE'", "WHEN 'FORM' THEN 0", "WHEN 'FORM_RECORD' THEN 1",
		"LIMIT 3", "FOR UPDATE",
	} {
		if !strings.Contains(listTaskResourceLinksForUpdateSQL, fragment) {
			t.Fatalf("resource link lock SQL missing %q", fragment)
		}
	}
}
