package postgres

import "testing"

func TestCutoverApplicationNamesFitPostgresLimit(t *testing.T) {
	const runID = "10000000-0000-4000-8000-000000000002"
	values := map[string]string{
		"adoption":     cutoverAdoptionApplicationName(runID),
		"finalization": cutoverFinalizerApplicationName(runID),
	}
	for name, value := range values {
		if len(value) > postgresApplicationNameLimit {
			t.Fatalf("%s application_name length=%d", name, len(value))
		}
	}
}
