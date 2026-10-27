package async

import "testing"

func analyticsDashboardExportMigrationPairs(t *testing.T) map[string]map[string][]byte {
	t.Helper()
	fresh := analyticsDashboardExportFreshMigrations(t)
	pairs := make(map[string]map[string][]byte, len(fresh))
	for name, document := range fresh {
		pairs[name] = map[string][]byte{"fresh": document}
	}
	return pairs
}
