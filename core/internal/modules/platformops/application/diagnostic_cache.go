package application

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/diagnostics"
)

func (reader *DiagnosticReader) cached(
	id api.DiagnosticID,
	now time.Time,
) (api.DiagnosticSnapshot, bool) {
	reader.mu.Lock()
	defer reader.mu.Unlock()
	snapshot, found := reader.cache[id]
	if !found || now.Before(snapshot.ObservedAt) ||
		now.Sub(snapshot.ObservedAt) > diagnostics.CacheTTL {
		return api.DiagnosticSnapshot{}, false
	}
	snapshot.Fields = append([]api.DiagnosticField(nil), snapshot.Fields...)
	return snapshot, true
}

func (reader *DiagnosticReader) store(snapshot api.DiagnosticSnapshot) {
	reader.mu.Lock()
	defer reader.mu.Unlock()
	snapshot.Fields = append([]api.DiagnosticField(nil), snapshot.Fields...)
	reader.cache[snapshot.ID] = snapshot
}
