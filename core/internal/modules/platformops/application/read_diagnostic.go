package application

import (
	"context"
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/diagnostics"
)

func (reader *DiagnosticReader) ReadDiagnostic(
	ctx context.Context,
	request api.DiagnosticRequest,
) (api.DiagnosticSnapshot, error) {
	if reader == nil || ctx == nil || ctx.Err() != nil {
		return api.DiagnosticSnapshot{}, diagnostics.ErrUnavailable
	}
	id := diagnostics.ID(request.ID)
	consumer := diagnostics.Consumer(request.Consumer)
	definition, err := diagnostics.Resolve(id)
	if err != nil || diagnostics.ValidateConsumer(consumer) != nil {
		return api.DiagnosticSnapshot{}, errors.Join(diagnostics.ErrUnavailable, err)
	}
	now := reader.clock.Now()
	if diagnostics.IsFailOpen(consumer) {
		if cached, found := reader.cached(request.ID, now); found {
			cached.Cached = true
			return cached, nil
		}
	}
	checkContext, cancel := context.WithTimeout(ctx, diagnostics.CheckTimeout)
	defer cancel()
	domainFields, err := reader.runner.RunDiagnostic(checkContext, definition)
	if err != nil {
		return reader.fallback(request, now, err)
	}
	if len(domainFields) > definition.MaximumRows || diagnosticBytes(domainFields) > definition.MaximumBytes {
		return reader.fallback(request, now, diagnostics.ErrUnavailable)
	}
	snapshot := api.DiagnosticSnapshot{ID: request.ID,
		OutputSchemaSHA256: definition.OutputSchemaSHA256,
		ObservedAt:         now, Fields: diagnosticAPIFields(domainFields)}
	reader.store(snapshot)
	return snapshot, nil
}

func diagnosticBytes(fields []diagnostics.Field) int {
	total := 0
	for _, field := range fields {
		total += len(field.Name) + len(field.Value)
	}
	return total
}

func diagnosticAPIFields(fields []diagnostics.Field) []api.DiagnosticField {
	result := make([]api.DiagnosticField, len(fields))
	for index, field := range fields {
		result[index] = api.DiagnosticField{Name: field.Name, Value: field.Value}
	}
	return result
}

func (reader *DiagnosticReader) fallback(
	request api.DiagnosticRequest,
	now time.Time,
	cause error,
) (api.DiagnosticSnapshot, error) {
	if diagnostics.IsFailOpen(diagnostics.Consumer(request.Consumer)) {
		if cached, found := reader.cached(request.ID, now); found {
			cached.Cached, cached.Stale = true, true
			return cached, nil
		}
	}
	return api.DiagnosticSnapshot{}, errors.Join(diagnostics.ErrUnavailable, cause)
}
