package application

import (
	"sync"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type DiagnosticReader struct {
	runner ports.DiagnosticRunner
	clock  ports.Clock
	mu     sync.Mutex
	cache  map[api.DiagnosticID]api.DiagnosticSnapshot
}

func NewDiagnosticReader(runner ports.DiagnosticRunner, clock ports.Clock) *DiagnosticReader {
	if runner == nil || clock == nil {
		return nil
	}
	return &DiagnosticReader{runner: runner, clock: clock,
		cache: make(map[api.DiagnosticID]api.DiagnosticSnapshot)}
}
