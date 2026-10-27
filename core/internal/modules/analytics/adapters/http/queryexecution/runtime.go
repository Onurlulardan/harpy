package queryexecutionhttp

import (
	"fmt"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

type Handler struct {
	runtime     *httptransport.Runtime
	application Application
}

func NewHandler(runtime *httptransport.Runtime, application Application) (*Handler, error) {
	if runtime == nil || httptransport.PortMissing(application) {
		return nil, fmt.Errorf("analytics query execution HTTP dependencies are required")
	}
	return &Handler{runtime: runtime, application: application}, nil
}
