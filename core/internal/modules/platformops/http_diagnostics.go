package platformops

import (
	"net/http"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type diagnosticFieldView struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type diagnosticSnapshotView struct {
	DiagnosticID       platformopsapi.DiagnosticID `json:"diagnostic_id"`
	OutputSchemaSHA256 string                      `json:"output_schema_sha256"`
	ObservedAt         time.Time                   `json:"observed_at"`
	Fields             []diagnosticFieldView       `json:"fields"`
	Cached             bool                        `json:"cached"`
	Stale              bool                        `json:"stale"`
}

func (handler *HTTPHandler) getDatabaseStats(w http.ResponseWriter, r *http.Request) {
	handler.readDiagnostic(w, r, platformopsapi.DiagnosticDatabaseStats,
		"platform.database.stats.read")
}

func (handler *HTTPHandler) getDatabaseSchema(w http.ResponseWriter, r *http.Request) {
	handler.readDiagnostic(w, r, platformopsapi.DiagnosticSchemaMetadata,
		"platform.database.schema.read")
}

func (handler *HTTPHandler) getRecoveryReadiness(w http.ResponseWriter, r *http.Request) {
	handler.readDiagnostic(w, r, platformopsapi.DiagnosticRecoveryReadiness,
		"platform.database.diagnostics.read")
}

func (handler *HTTPHandler) readDiagnostic(
	w http.ResponseWriter, r *http.Request,
	id platformopsapi.DiagnosticID, capability string,
) {
	environment, err := handler.environmentFor(r)
	if err == nil {
		_, err = handler.accessFor(r, environment, capability, "")
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	snapshot, err := handler.diagnostics.ReadDiagnostic(r.Context(), platformopsapi.DiagnosticRequest{
		ID: id, Consumer: platformopsapi.ConsumerInteractive})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	fields := make([]diagnosticFieldView, len(snapshot.Fields))
	for index, field := range snapshot.Fields {
		fields[index] = diagnosticFieldView{Name: field.Name, Value: field.Value}
	}
	handler.writeSuccess(w, r, http.StatusOK, diagnosticSnapshotView{
		DiagnosticID: snapshot.ID, OutputSchemaSHA256: snapshot.OutputSchemaSHA256,
		ObservedAt: snapshot.ObservedAt,
		Fields:     fields, Cached: snapshot.Cached, Stale: snapshot.Stale})
}
