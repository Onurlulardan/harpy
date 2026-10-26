package worker

import (
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func TestWorkJobCompositionContract(t *testing.T) {
	var builder func(platformasync.ReleaseCatalogs, asyncpostgres.WorkerStore) (
		[]jobexecution.Processor, error,
	) = buildWorkJobProcessors
	if builder == nil {
		t.Fatal("Work processor builder is not linked")
	}
}
