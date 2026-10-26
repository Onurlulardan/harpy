package cataloghttp

import (
	"encoding/json"
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

const moduleMigrationSuccessor = "/api/catalog/installations/{installation_id}/upgrades"

func ModuleMigrationRetirementRoutes(runtime *Runtime) []Route {
	if !runtime.Ready() {
		return nil
	}
	retired := http.HandlerFunc(runtime.retiredModuleMigration)
	return []Route{
		{Method: http.MethodGet,
			Pattern: "/api/form-engine/modules/{slug}/migrations/plan", Handler: retired},
		{Method: http.MethodPost,
			Pattern: "/api/form-engine/modules/{slug}/migrations/apply", Handler: retired},
		{Method: http.MethodGet,
			Pattern: "/api/form-engine/modules/{slug}/migrations/runs", Handler: retired},
		{Method: http.MethodGet,
			Pattern: "/api/form-engine/modules/{slug}/migrations/runs/{run_id}/field-backups",
			Handler: retired},
	}
}

func (runtime *Runtime) retiredModuleMigration(
	writer http.ResponseWriter,
	request *http.Request,
) {
	writer.Header().Set("Link", "<"+moduleMigrationSuccessor+">; rel=\"successor-version\"")
	correlation := runtime.correlation(request)
	envelope, err := runtimev1.NewErrorEnvelope(
		"STANDALONE_MODULE_MIGRATION_RETIRED",
		"Standalone module migration is retired; use the Catalog upgrade plan and receipt",
		correlation, nil,
	)
	var body []byte
	if err == nil {
		body, err = json.Marshal(envelope)
	}
	if err == nil {
		err = writeJSON(writer, http.StatusGone, correlation, body)
	}
	if err != nil {
		runtime.report(errors.New("catalog module migration retirement response failed"))
	}
}
