//go:build integration

package integrationtest

import (
	"path"
	"sort"
	"strings"
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformmigration "github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func analyticsDashboardExportOrigin(t *testing.T) string {
	t.Helper()
	return "fresh"
}

func analyticsDashboardExportFoundationDocuments(t *testing.T,
	origin string,
) []platformmigration.Document {
	t.Helper()
	wanted := map[string]bool{
		"20260715020000": true, "20260715020100": true,
		"20260715020200": true, "20260715020300": true,
		"20260715020400": true, "20260715020500": true,
		"20260715020600": true, "20260715020700": true,
		"20260715020800": true, "20260715020820": true,
		"20260715020900": true, "20260715021600": true,
		"20260715021200": true, "20260715021250": true,
		"20260715021510": true, "20260715021520": true,
		"20260715021530": true, "20260715021540": true,
		"20260715021542": true, "20260715021544": true,
		"20260715021555": true, "20260715021556": true,
		"20260715021557": true, "20260715021558": true,
		"20260715021559": true,
		"20260802020001": true,
		"20260802020007": true, "20260802020014": true,
		"20260802020021": true,
	}
	return selectAnalyticsDashboardExportPlatformDocuments(t, origin, wanted)
}

func analyticsDashboardExportProtocolDocuments(t *testing.T, origin,
	first, last string,
) []platformmigration.Document {
	t.Helper()
	wanted := make(map[string]bool)
	for suffix := 0; suffix <= 49; suffix++ {
		stamp := "202608020200" + twoDigitAnalyticsExportStamp(suffix)
		if stamp >= first && stamp <= last &&
			stamp != "20260802020001" && stamp != "20260802020007" &&
			stamp != "20260802020014" && stamp != "20260802020021" {
			wanted[stamp] = true
		}
	}
	return selectAnalyticsDashboardExportPlatformDocuments(t, origin, wanted)
}

func twoDigitAnalyticsExportStamp(value int) string {
	return string([]byte{'0' + byte(value/10), '0' + byte(value%10)})
}

func selectAnalyticsDashboardExportPlatformDocuments(t *testing.T, origin string,
	wanted map[string]bool,
) []platformmigration.Document {
	t.Helper()
	if origin != "fresh" && origin != "v1_adopted" {
		t.Fatalf("Platform Analytics export origin %q is unsupported", origin)
	}
	documents, err := platformasync.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	selected := make([]platformmigration.Document, 0, len(wanted))
	for _, document := range documents {
		name := path.Base(document.Path())
		stamp := strings.SplitN(name, "_", 2)[0]
		if strings.Contains(document.Path(), "/fresh/") && wanted[stamp] {
			selected = append(selected, document)
			delete(wanted, stamp)
		}
	}
	if len(wanted) != 0 {
		t.Fatalf("Platform Analytics export %s migrations missing: %v", origin, wanted)
	}
	sort.Slice(selected, func(left, right int) bool {
		return selected[left].Path() < selected[right].Path()
	})
	return selected
}
