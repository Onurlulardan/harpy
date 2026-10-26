package compatibilityhttp

import (
	"fmt"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

type Handler struct {
	runtime     *httptransport.Runtime
	application Application
	public      PublicApplication
}

func NewHandler(
	runtime *httptransport.Runtime, application Application, public PublicApplication,
) (*Handler, error) {
	if runtime == nil || httptransport.PortMissing(application) || httptransport.PortMissing(public) {
		return nil, fmt.Errorf("analytics compatibility HTTP dependencies are required")
	}
	return &Handler{runtime: runtime, application: application, public: public}, nil
}
