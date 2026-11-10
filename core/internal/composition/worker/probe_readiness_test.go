package worker

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/health"
)

func TestReadinessFailureNamesTheFailedCheck(t *testing.T) {
	cases := []struct {
		states health.States
		want   string
	}{
		{health.States{}, "database readiness probe failed"},
		{health.States{Database: true}, "database role readiness probe failed"},
		{health.States{Database: true, DatabaseRole: true},
			"deployment policy readiness probe failed"},
	}
	for _, test := range cases {
		if got := readinessFailure(test.states); got != test.want {
			t.Fatalf("readinessFailure(%+v) = %s", test.states, got)
		}
	}
}
